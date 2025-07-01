package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"todoai/internal/config"
	"todoai/internal/gateway/ai"
	"todoai/pkg/reader/docx"
	"todoai/pkg/reader/pdf"
)

const (
	bookMode          = "book"
	docMode           = "doc"
	articleMode       = "article"
	jurisprudenceMode = "jurisprudence"
	defaultMode       = "default"
)

var ErrFileTooLarge = errors.New("the file is too large")

type textExtractor func(fileBytes []byte) (string, error)

type Upload interface {
	ProcessFile(ctx context.Context, fileBytes []byte, filename string, mode string) (string, error)
	ProcessText(ctx context.Context, text string, mode string) (string, error)
}

type file struct {
	config     *config.Config
	log        *slog.Logger
	ai         ai.AI
	extractors map[string]textExtractor
}

type processingInstructions struct {
	CompressorInstruction string
	FormatterInstruction  string
}

func NewFileService(config *config.Config, log *slog.Logger, ai ai.AI) *file {
	s := &file{
		config: config,
		log:    log,
		ai:     ai,
	}

	s.extractors = map[string]textExtractor{
		".txt":  s.extractFromPlainText,
		".csv":  s.extractFromPlainText,
		".log":  s.extractFromPlainText,
		".md":   s.extractFromPlainText,
		".pdf":  s.extractFromPDF,
		".docx": s.extractFromDocx,
	}

	return s
}

func (s *file) ProcessFile(ctx context.Context, fileBytes []byte, filename string, mode string) (string, error) {
	const op = "service.ProcessFile"
	fileExt := strings.ToLower(filepath.Ext(filename))
	extractor, ok := s.extractors[fileExt]
	if !ok {
		s.log.Error(op, "error", fmt.Errorf("unsupported file type: %s", fileExt))
		return "", errors.New("unsupported file type")
	}
	text, err := extractor(fileBytes)
	if err != nil {
		s.log.Error(op, "error", err)
		return "", err
	}
	instructions, err := s.getInstructions(mode)
	if err != nil {
		s.log.Error(op, "error", err)
		return "", err
	}
	return s.compress(ctx, text, instructions)
}

func (s *file) ProcessText(ctx context.Context, text string, mode string) (string, error) {
	const op = "service.ProcessText"
	instructions, err := s.getInstructions(mode)
	if err != nil {
		s.log.Error(op, "error", err)
		return "", err
	}
	return s.compress(ctx, text, instructions)
}

func (s *file) compress(ctx context.Context, fileText string, instructions processingInstructions) (string, error) {
	const op = "service.Compress"
	chunks := s.spliter(fileText)
	if len(chunks) > s.config.Summarizer.MaxLen {
		return "", ErrFileTooLarge
	}
	var content strings.Builder
	for _, v := range chunks {
		part, err := s.ai.Generate(ctx, v, instructions.CompressorInstruction, s.config.Summarizer.CompressTemp)
		if err != nil {
			s.log.Error(op, "error", err)
			return "", err
		}
		content.WriteString(fmt.Sprintf("%s\n", part))
	}

	finalText, err := s.ai.Generate(ctx, content.String(), instructions.FormatterInstruction, s.config.Summarizer.FormatTemp)
	if err != nil {
		s.log.Error(op, "error", err)
		return "", err
	}
	return finalText, nil
}

func (s *file) extractFromPlainText(fileBytes []byte) (string, error) {
	return string(fileBytes), nil
}

func (s *file) extractFromPDF(fileBytes []byte) (string, error) {
	const op = "sdervice.ExtractFromPDF"
	text, err := pdf.ReadPDF(fileBytes)
	if err != nil {
		s.log.Error(op, "error", err)
		return "", err
	}
	return text, nil
}

func (s *file) extractFromDocx(fileBytes []byte) (string, error) {
	const op = "service.ExtractFromDocx"
	text, err := docx.ReadDocx(fileBytes)
	if err != nil {
		s.log.Error(op, "error", err)
		return "", err
	}
	return text, nil
}

func (s *file) getInstructions(mode string) (processingInstructions, error) {
	switch mode {
	case bookMode:
		return processingInstructions{
			CompressorInstruction: s.config.Instructions.Book,
			FormatterInstruction:  s.config.Instructions.Formatter_book,
		}, nil
	case jurisprudenceMode:
		return processingInstructions{
			CompressorInstruction: s.config.Instructions.Jurisprudence,
			FormatterInstruction:  s.config.Instructions.Formatter_jurisprudence,
		}, nil
	case articleMode:
		return processingInstructions{
			CompressorInstruction: s.config.Instructions.Article,
			FormatterInstruction:  s.config.Instructions.Formatter_article,
		}, nil
	case docMode:
		return processingInstructions{
			CompressorInstruction: s.config.Instructions.Doc,
			FormatterInstruction:  s.config.Instructions.Formatter_doc,
		}, nil
	case defaultMode:
		return processingInstructions{
			CompressorInstruction: s.config.Instructions.Def,
			FormatterInstruction:  s.config.Instructions.Formatter_def,
		}, nil
	default:
		return processingInstructions{}, errors.New("unsupported mode")
	}
}

func (s *file) spliter(text string) []string {
	parts := strings.Fields(text)
	var result []string

	chunkSize := s.config.Summarizer.ChunkSize
	for i := 0; i < len(parts); i += chunkSize {
		end := i + chunkSize
		if end > len(parts) {
			end = len(parts)
		}
		chunk := strings.Join(parts[i:end], " ")
		result = append(result, chunk)
	}

	return result
}
