package output

import (
	"fmt"
	"time"

	"github.com/grokspawn/executive-brief/internal/config"
	"github.com/grokspawn/executive-brief/internal/matrix"
)

// GenerateHTML generates an HTML executive brief
func GenerateHTML(items *matrix.CategorizedItems, cfg *config.Config, startTime, endTime time.Time) string {
	// For now, just return markdown wrapped in HTML
	// Can be enhanced later with proper HTML/CSS layout
	markdown := GenerateMarkdown(items, cfg, startTime, endTime)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Executive Brief - %s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
            max-width: 1200px;
            margin: 40px auto;
            padding: 0 20px;
            line-height: 1.6;
        }
        h1 { color: #333; }
        h2 { color: #0366d6; border-bottom: 2px solid #e1e4e8; padding-bottom: 8px; }
        h3 { color: #586069; }
        ul { list-style: none; padding: 0; }
        li { margin: 15px 0; padding: 10px; border-left: 4px solid #e1e4e8; }
        a { color: #0366d6; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .quadrant-1 { border-left-color: #d73a49; }
        .quadrant-2 { border-left-color: #f9c513; }
        .quadrant-3 { border-left-color: #0366d6; }
        .quadrant-4 { border-left-color: #6a737d; }
        code { background: #f6f8fa; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
<pre>%s</pre>
</body>
</html>`, time.Now().Format("2006-01-02"), markdown)

	return html
}
