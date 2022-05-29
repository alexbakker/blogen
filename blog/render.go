package blog

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/russross/blackfriday/v2"
	yaml "gopkg.in/yaml.v3"
)

func (b *Blog) renderPost(post *Post, input []byte) error {
	var tocBuf bytes.Buffer
	var bodyBuf bytes.Buffer
	var sumBuf bytes.Buffer
	var sumText string

	renderer := blackfriday.NewHTMLRenderer(
		blackfriday.HTMLRendererParameters{
			Flags: blackfriday.CommonHTMLFlags,
		},
	)
	parser := blackfriday.New(
		blackfriday.WithRenderer(renderer),
		blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.AutoHeadingIDs|blackfriday.Footnotes),
	)
	ast := parser.Parse(input)

	renderTOC(renderer, &tocBuf, ast)

	var bodyErr error
	var foundInfo bool
	var foundSum bool
	var foundTitle bool
	var sumNode *blackfriday.Node

	// parse post info, render summary and render body
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		var skipSum bool

		switch node.Type {
		case blackfriday.CodeBlock:
			if entering {
				if !foundInfo {
					// parse post info
					if err := yaml.Unmarshal(node.Literal, post); err != nil {
						bodyErr = err
						return blackfriday.Terminate
					}
					foundInfo = true
				} else {
					// syntax-highlight any code blocks
					if err := b.renderCode(&bodyBuf, node.Literal, node.CodeBlockData); err != nil {
						bodyErr = err
						return blackfriday.Terminate
					}
				}
				return blackfriday.SkipChildren
			}
		case blackfriday.Paragraph:
			// use the first paragraph as a summary for the post
			if !foundSum {
				if entering && sumNode == nil {
					sumNode = node
					skipSum = true
				}
				if !entering && sumNode == node {
					sumNode = nil
					foundSum = true
					skipSum = true
				}
			}
		case blackfriday.Heading:
			// don't render the title here
			if node.HeadingData.Level == 1 {
				return blackfriday.GoToNext
			}
		case blackfriday.Text:
			// take out the title
			if entering && !foundTitle && node.Parent != nil && node.Parent.Type == blackfriday.Heading {
				foundTitle = true
				return blackfriday.SkipChildren
			}

			// store summary text
			if entering && !foundSum && node.Parent != nil {
				sumText += strings.Replace(string(node.Literal), "\n", " ", -1)
			}
		}

		if sumNode != nil && !skipSum {
			status := renderer.RenderNode(&sumBuf, node, entering)
			if status == blackfriday.Terminate {
				return blackfriday.Terminate
			}
		}

		return renderer.RenderNode(&bodyBuf, node, entering)
	})

	if bodyErr != nil {
		return bodyErr
	}
	if !foundInfo {
		return errors.New("post info not found")
	}
	if !foundSum {
		return errors.New("couldn't extract post summary")
	}

	renderer.RenderFooter(&bodyBuf, ast)

	post.TOC = template.HTML(tocBuf.Bytes())
	post.Content = template.HTML(bodyBuf.Bytes())
	post.Summary = template.HTML(sumBuf.Bytes())
	post.SummaryText = sumText
	return nil
}

func (b *Blog) renderCode(w io.Writer, literal []byte, data blackfriday.CodeBlockData) error {
	lang := string(data.Info)
	text := string(bytes.TrimRight(literal, "\n"))

	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Analyse(text)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	codeStyle := styles.Get(b.theme.Style.Syntax.Name)
	if codeStyle == nil {
		return errors.New("style not found")
	}

	iterator, err := lexer.Tokenise(nil, text)
	if err != nil {
		return err
	}

	formatter := html.New(html.WithClasses(true), html.WithLineNumbers(b.theme.Style.Syntax.Numbered), html.LineNumbersInTable(true))
	return formatter.Format(w, codeStyle, iterator)
}

// copied from the blackfriday source and modified to exclude the title of the post
func renderTOC(r *blackfriday.HTMLRenderer, w io.Writer, ast *blackfriday.Node) {
	buf := bytes.Buffer{}

	inHeading := false
	tocLevel := 0
	headingCount := 0

	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if node.Type == blackfriday.Heading && !node.HeadingData.IsTitleblock {
			if node.HeadingData.Level == 1 {
				return blackfriday.GoToNext
			}
			level := node.Level - 1

			inHeading = entering
			if entering {
				if node.HeadingID == "" {
					node.HeadingID = fmt.Sprintf("toc_%d", headingCount)
				}
				if level == tocLevel {
					buf.WriteString("</li>\n\n<li>")
				} else if level < tocLevel {
					for level < tocLevel {
						tocLevel--
						buf.WriteString("</li>\n</ul>")
					}
					buf.WriteString("</li>\n\n<li>")
				} else {
					for level > tocLevel {
						tocLevel++
						buf.WriteString("\n<ul>\n<li>")
					}
				}

				fmt.Fprintf(&buf, `<a href="#%s">`, node.HeadingID)
				headingCount++
			} else {
				buf.WriteString("</a>")
			}
			return blackfriday.GoToNext
		}

		if inHeading {
			return r.RenderNode(&buf, node, entering)
		}

		return blackfriday.GoToNext
	})

	for ; tocLevel > 0; tocLevel-- {
		buf.WriteString("</li>\n</ul>")
	}

	if buf.Len() > 0 {
		io.WriteString(w, "<nav>\n")
		w.Write(buf.Bytes())
		io.WriteString(w, "\n\n</nav>\n")
	}
	//r.lastOutputLen = buf.Len()
}
