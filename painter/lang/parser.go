package lang

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"strconv"
	"strings"

	"github.com/roman-mazur/architecture-lab-3/painter"
)

// Parser handles parsing of drawing commands from text input
type Parser struct {
	lastBgColor painter.Operation
	lastBgRect  *painter.BgRectangle
	figures     []*painter.Figure
	moveOps     []painter.Operation
	updateOp    painter.Operation
}

// initialize sets up the initial state for parsing
func (p *Parser) initialize() {
	if p.lastBgColor == nil {
		p.lastBgColor = painter.OperationFunc(painter.ResetScreen)
	}
	if p.updateOp != nil {
		p.updateOp = nil
	}
}

func (p *Parser) Parse(in io.Reader) ([]painter.Operation, error) {
	if in == nil {
		return nil, fmt.Errorf("input reader is nil")
	}

	p.initialize()
	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		commandLine := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if commandLine == "" || strings.HasPrefix(commandLine, "#") {
			continue
		}

		err := p.parse(commandLine)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return p.finalResult(), nil
}

// finalResult builds the final list of operations
func (p *Parser) finalResult() []painter.Operation {
	// Pre-allocate the slice to avoid reallocations
	totalOps := 0
	if p.lastBgColor != nil {
		totalOps++
	}
	if p.lastBgRect != nil {
		totalOps++
	}
	totalOps += len(p.moveOps) + len(p.figures)
	if p.updateOp != nil {
		totalOps++
	}

	res := make([]painter.Operation, 0, totalOps)

	if p.lastBgColor != nil {
		res = append(res, p.lastBgColor)
	}
	if p.lastBgRect != nil {
		res = append(res, p.lastBgRect)
	}
	if len(p.moveOps) != 0 {
		res = append(res, p.moveOps...)
	}
	// Reset moveOps to avoid memory leaks
	p.moveOps = nil

	if len(p.figures) != 0 {
		for _, figure := range p.figures {
			res = append(res, figure)
		}
	}
	if p.updateOp != nil {
		res = append(res, p.updateOp)
	}
	return res
}

// resetState resets the parser's state
func (p *Parser) resetState() {
	p.lastBgColor = nil
	p.lastBgRect = nil
	p.figures = nil
	p.moveOps = nil
	p.updateOp = nil
}

// parseIntArgs parses string arguments into integers
func parseIntArgs(args []string) ([]int, error) {
	result := make([]int, 0, len(args))
	for i, arg := range args {
		val, err := strconv.Atoi(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid integer argument at position %d: %s", i, arg)
		}
		result = append(result, val)
	}
	return result, nil
}

// parse processes a single command line
func (p *Parser) parse(commandLine string) error {
	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		return nil
	}

	instruction := parts[0]
	args := parts[1:]

	switch instruction {
	case "white":
		p.lastBgColor = painter.OperationFunc(painter.WhiteFill)
		return nil

	case "green":
		p.lastBgColor = painter.OperationFunc(painter.GreenFill)
		return nil

	case "bgrect":
		if len(args) < 4 {
			return fmt.Errorf("bgrect command requires 4 arguments, got %d", len(args))
		}

		iArgs, err := parseIntArgs(args[:4])
		if err != nil {
			return fmt.Errorf("invalid bgrect arguments: %w", err)
		}

		p.lastBgRect = &painter.BgRectangle{X1: iArgs[0], Y1: iArgs[1], X2: iArgs[2], Y2: iArgs[3]}
		return nil

	case "figure":
		if len(args) < 2 {
			return fmt.Errorf("figure command requires at least 2 arguments, got %d", len(args))
		}

		iArgs, err := parseIntArgs(args[:2])
		if err != nil {
			return fmt.Errorf("invalid figure arguments: %w", err)
		}

		clr := color.RGBA{B: 255, A: 255}
		figure := &painter.Figure{X: iArgs[0], Y: iArgs[1], C: clr}
		p.figures = append(p.figures, figure)
		return nil

	case "move":
		if len(args) < 2 {
			return fmt.Errorf("move command requires 2 arguments, got %d", len(args))
		}

		iArgs, err := parseIntArgs(args[:2])
		if err != nil {
			return fmt.Errorf("invalid move arguments: %w", err)
		}

		var figuresCopy []*painter.Figure
		if len(p.figures) > 0 {
			figuresCopy = make([]*painter.Figure, len(p.figures))
			copy(figuresCopy, p.figures)
		}

		moveOp := &painter.Move{X: iArgs[0], Y: iArgs[1], Figures: figuresCopy}
		p.moveOps = append(p.moveOps, moveOp)
		return nil

	case "reset":
		p.resetState()
		p.lastBgColor = painter.OperationFunc(painter.ResetScreen)
		return nil

	case "update":
		p.updateOp = painter.UpdateOp
		return nil

	default:
		return fmt.Errorf("unknown command: %s", instruction)
	}
}
