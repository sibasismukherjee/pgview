package clipboard

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Write copies text to the system clipboard.
//
// Inside a tmux session ($TMUX is set) it uses tmux set-buffer, which works
// on any OS without requiring X11 or Wayland forwarding.
// Outside tmux: pbcopy on macOS; wl-copy, xclip, or xsel on Linux.
func Write(text string) error {
	if os.Getenv("TMUX") != "" {
		return exec.Command("tmux", "set-buffer", "--", text).Run()
	}
	switch runtime.GOOS {
	case "darwin":
		return pipeCmd(text, "pbcopy")
	default:
		for _, args := range [][]string{
			{"wl-copy"},
			{"xclip", "-selection", "clipboard"},
			{"xsel", "--clipboard", "--input"},
		} {
			if pipeCmd(text, args[0], args[1:]...) == nil {
				return nil
			}
		}
		return fmt.Errorf("no clipboard tool found (tried wl-copy, xclip, xsel)")
	}
}

func pipeCmd(text, name string, args ...string) error {
	if _, err := exec.LookPath(name); err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
