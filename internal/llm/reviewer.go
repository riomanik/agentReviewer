package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature"`
}

func Review(prompt string) (string, error) {

	reqBody := Request{
		Model:       os.Getenv("LLM_MODEL"),
		MaxTokens:   1500,
		Temperature: 0.2,
		Messages: []Message{
			{
				Role: "system",
				Content: `
Kamu adalah Senior Software Engineer yang berpengalaman melakukan code review di lingkungan production.

Tugas kamu:
- Melakukan review terhadap perubahan kode pada sebuah Pull Request
- Mengidentifikasi bug potensial, kesalahan logika, dan edge case
- Memberikan masukan terkait kualitas kode, readability, performa, dan maintainability
- Menunjukkan pelanggaran best practice (jika ada)
- Memberikan saran perbaikan yang konkret dan masuk akal

Aturan:
- Gunakan Bahasa Indonesia yang profesional, ringkas, dan jelas
- Bersikap objektif dan konstruktif
- Referensikan nama file dan potongan kode jika relevan
- Jangan mengulang isi diff
- Jangan berasumsi di luar konteks yang diberikan
- Jika tidak ada masalah kritis, tuliskan: "Tidak ditemukan masalah kritis"

Format output (WAJIB):
Ringkasan:
Masalah Kritis:
Saran Perbaikan:
Pertanyaan untuk Author (opsional):
`,
			},
			{Role: "user", Content: prompt},
		},
	}

	data, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(
		"POST",
		os.Getenv("LLM_BASE_URL")+"/chat/completions",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("LLM_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	// 🔥 WAJIB UNTUK OPENROUTER
	req.Header.Set("HTTP-Referer", "https://github.com/yourname/ai-reviewer")
	req.Header.Set("X-Title", "AI PR Reviewer Bot")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 🔥 DEBUG + SAFETY
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"LLM error status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("LLM returned empty choices")
	}

	return result.Choices[0].Message.Content, nil
}
