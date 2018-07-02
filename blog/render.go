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

func (b *Blog) renderPost(info *PostInfo, input []byte) (*Post, error) {
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
		var skip bool
		var skipSum bool

		switch node.Type {
		case blackfriday.CodeBlock:
			if entering {
				if !foundInfo && string(node.CodeBlockData.Info) == "post" {
					// parse post info
					if err := yaml.Unmarshal(node.Literal, info); err != nil {
						bodyErr = err
						return blackfriday.Terminate
					}
					foundInfo = true
					skip = true
				} else {
					// syntax-highlight any code blocks
					var codeBuf bytes.Buffer
					if err := b.renderCode(&codeBuf, string(node.Literal), string(node.CodeBlockData.Info)); err != nil {
						bodyErr = err
						return blackfriday.Terminate
					}

					node = blackfriday.NewNode(blackfriday.HTMLBlock)
					node.Literal = codeBuf.Bytes()
				}
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
				sumText += strings.TrimSpace(strings.Replace(string(node.Literal), "\n", " ", -1))
			}
		}

		if sumNode != nil && !skipSum {
			status := renderer.RenderNode(&sumBuf, node, entering)
			if status == blackfriday.Terminate {
				return blackfriday.Terminate
			}
		}

		if !skip {
			return renderer.RenderNode(&bodyBuf, node, entering)
		}

		return blackfriday.GoToNext
	})

	if bodyErr != nil {
		return nil, bodyErr
	}
	if !foundInfo {
		return nil, errors.New("post info not found")
	}
	if !foundSum {
		return nil, errors.New("couldn't extract post summary")
	}

	// render footer
	renderer.RenderFooter(&bodyBuf, ast)

	info.Summary = template.HTML(sumBuf.Bytes())
	info.SummaryText = sumText

	return &Post{
		Info:    info,
		Content: template.HTML(bodyBuf.Bytes()),
	}, nil
}

func (b *Blog) renderCode(output io.Writer, input string, lang string) error {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Analyse(input)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	codeStyle := styles.Get(b.config.CodeStyle)
	if codeStyle == nil {
		return errors.New("style not found")
	}

	iterator, err := lexer.Tokenise(nil, input)
	if err != nil {
		return err
	}

	formatter := html.New(html.WithClasses(), html.WithLineNumbers(), html.LineNumbersInTable())
	return formatter.Format(output, codeStyle, iterator)
}
