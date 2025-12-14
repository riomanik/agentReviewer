package output

import "fmt"

func ToMarkdown(repo string, pr int, content string) string {
	return fmt.Sprintf(`# 🧠 AI Code Review

**Repository:** %s  
**Pull Request:** #%d  

---

%s
`, repo, pr, content)
}
