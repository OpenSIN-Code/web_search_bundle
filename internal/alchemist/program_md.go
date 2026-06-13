// Purpose: Parse and update program.md research plans.
// Docs: program_md.doc.md

package alchemist

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// ProgramMD represents the agent's self-updating research plan
type ProgramMD struct {
	path       string
	hypotheses []string // pending queue
	learnings  []string // accumulated knowledge
	rawContent string
	mu         sync.Mutex
}

// LoadProgramMD parses a program.md file
func LoadProgramMD(path string) (*ProgramMD, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p := &ProgramMD{
		path:       path,
		rawContent: string(data),
	}
	p.parse()
	return p, nil
}

// parse extracts hypotheses and learnings from markdown
func (p *ProgramMD) parse() {
	scanner := bufio.NewScanner(strings.NewReader(p.rawContent))
	var section string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect sections
		lower := strings.ToLower(trimmed)
		switch {
		case strings.Contains(lower, "hypothesis") && strings.Contains(lower, "queue"):
			section = "hypotheses"
			continue
		case strings.Contains(lower, "learning"):
			section = "learnings"
			continue
		case strings.HasPrefix(trimmed, "#"):
			section = ""
			continue
		}

		// Parse items
		switch section {
		case "hypotheses":
			if strings.HasPrefix(trimmed, "- [ ]") {
				h := strings.TrimSpace(strings.TrimPrefix(trimmed, "- [ ]"))
				if h != "" {
					p.hypotheses = append(p.hypotheses, h)
				}
			}
		case "learnings":
			if strings.HasPrefix(trimmed, "-") {
				l := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
				l = strings.TrimPrefix(l, "*")
				l = strings.TrimSpace(l)
				if l != "" {
					p.learnings = append(p.learnings, l)
				}
			}
		}
	}
}

// NextHypothesis returns and removes the next pending hypothesis
func (p *ProgramMD) NextHypothesis() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.hypotheses) == 0 {
		return ""
	}

	h := p.hypotheses[0]
	p.hypotheses = p.hypotheses[1:]

	// Mark as in-progress in the file (non-blocking)
	go p.markHypothesisInProgress(h)

	return h
}

// AppendLearning adds a new learning to the queue and persists to disk
func (p *ProgramMD) AppendLearning(learning string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.learnings = append(p.learnings, learning)

	// Persist to disk (best-effort)
	if err := p.flush(); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ program.md flush failed: %v\n", err)
	}
}

// Hypotheses returns pending hypotheses (copy)
func (p *ProgramMD) Hypotheses() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]string, len(p.hypotheses))
	copy(out, p.hypotheses)
	return out
}

// Learnings returns accumulated learnings (copy)
func (p *ProgramMD) Learnings() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]string, len(p.learnings))
	copy(out, p.learnings)
	return out
}

// AddHypothesis adds a new hypothesis to the queue
func (p *ProgramMD) AddHypothesis(h string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.hypotheses = append(p.hypotheses, h)
	_ = p.flush()
}

// flush writes the updated program.md to disk
func (p *ProgramMD) flush() error {
	// Build updated content: preserve everything, update learnings section
	lines := strings.Split(p.rawContent, "\n")
	var out []string

	inLearnings := false
	learningsWritten := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		// Detect learnings section start
		if strings.Contains(lower, "learning") && strings.HasPrefix(trimmed, "#") {
			inLearnings = true
			out = append(out, line)
			// Write existing learnings right after the header
			for _, l := range p.learnings {
				out = append(out, "- "+l)
			}
			learningsWritten = true
			continue
		}

		// Detect next section (end of learnings)
		if inLearnings && strings.HasPrefix(trimmed, "#") {
			inLearnings = false
		}

		// Skip old learning bullets (we rewrote them)
		if inLearnings && strings.HasPrefix(trimmed, "-") {
			continue
		}

		out = append(out, line)
	}

	// If no learnings section found, append one
	if !learningsWritten {
		out = append(out, "", "## Learnings (agent updates this)", "")
		for _, l := range p.learnings {
			out = append(out, "- "+l)
		}
	}

	content := strings.Join(out, "\n")
	return os.WriteFile(p.path, []byte(content), 0644)
}

// markHypothesisInProgress marks a hypothesis as [~] in the file
func (p *ProgramMD) markHypothesisInProgress(h string) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		return
	}

	content := string(data)
	oldMark := "- [ ] " + h
	newMark := "- [~] " + h
	content = strings.Replace(content, oldMark, newMark, 1)

	_ = os.WriteFile(p.path, []byte(content), 0644)
}
