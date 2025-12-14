package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"agentReviewer/internal/llm"
	"agentReviewer/internal/output"
	"agentReviewer/internal/pr"

	"github.com/bwmarrin/discordgo"
)

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		panic("DISCORD_BOT_TOKEN not set")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	dg.AddHandler(onMessageCreate)

	err = dg.Open()
	if err != nil {
		panic(err)
	}

	fmt.Println("🤖 Discord bot is running")
	select {} // block forever
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// ignore bot message
	if m.Author.Bot {
		return
	}

	// command format
	// !review https://github.com/org/repo/pull/123
	if !strings.HasPrefix(m.Content, "!review ") {
		return
	}

	_, _ = s.ChannelMessageSend(
		m.ChannelID,
		"⏳ Sedang melakukan review PR, mohon tunggu...",
	)

	prURL := strings.TrimSpace(strings.TrimPrefix(m.Content, "!review "))

	// 1️⃣ PARSE PR URL
	info, err := pr.ParsePRURL(prURL)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "❌ PR URL tidak valid")
		return
	}

	// 2️⃣ FETCH PR FILES
	files, err := pr.FetchPRFiles(info)
	if err != nil {
		s.ChannelMessageSend(
			m.ChannelID,
			"❌ Gagal mengambil PR: "+err.Error(),
		)
		return
	}

	// 3️⃣ BUILD DIFF PROMPT
	var diffBuilder strings.Builder

	for _, f := range files {
		if f.Patch == "" {
			continue
		}

		diffBuilder.WriteString("\n---\n")
		diffBuilder.WriteString("File: " + f.Filename + "\n")
		diffBuilder.WriteString(f.Patch + "\n")
	}

	if diffBuilder.Len() == 0 {
		s.ChannelMessageSend(
			m.ChannelID,
			"⚠️ Tidak ada diff yang bisa direview (file binary / terlalu besar)",
		)
		return
	}

	prompt := fmt.Sprintf(
		"Repository: %s/%s\nPull Request: #%d\n\n%s",
		info.Owner,
		info.Repo,
		info.Number,
		diffBuilder.String(),
	)

	// 4️⃣ CALL LLM
	review, err := llm.Review(prompt)
	if err != nil {
		s.ChannelMessageSend(
			m.ChannelID,
			"❌ AI reviewer error",
		)
		return
	}

	// 5️⃣ GENERATE MARKDOWN
	md := output.ToMarkdown(
		fmt.Sprintf("%s/%s", info.Owner, info.Repo),
		info.Number,
		review,
	)

	// 6️⃣ SEND FILE TO DISCORD
	_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: "✅ Review selesai",
		Files: []*discordgo.File{
			{
				Name:        fmt.Sprintf("review-pr-%d.md", info.Number),
				ContentType: "text/markdown",
				Reader:      bytes.NewReader([]byte(md)),
			},
		},
	})

	if err != nil {
		fmt.Println("failed send discord message:", err)
	}
}
