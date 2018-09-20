package blog

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	blackfriday "gopkg.in/russross/blackfriday.v2"
	yaml "gopkg.in/yaml.v2"
)

func (b *Blog) renderPost(post *Post, input []byte) error {
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
		blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.AutoHeadingIDs),
	)
	ast := parser.Parse(input)

	// render header
	renderer.RenderHeader(&bodyBuf, ast)

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

	// render footer
	renderer.RenderFooter(&bodyBuf, ast)

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

	codeStyle := styles.Get(b.theme.Style.Syntax)
	if codeStyle == nil {
		return errors.New("style not found")
	}

	iterator, err := lexer.Tokenise(nil, text)
	if err != nil {
		return err
	}

	formatter := html.New(html.WithClasses(), html.WithLineNumbers(), html.LineNumbersInTable())
	return formatter.Format(w, codeStyle, iterator)
}
