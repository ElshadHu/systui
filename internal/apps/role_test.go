package apps

import "testing"

func TestRoleLabel(t *testing.T) {
	cases := []struct{ app, name, cmdline, want string }{
		{"Google Chrome", "Google Chrome Helper (Renderer)", "x --type=renderer", "Renderer"},
		{"Google Chrome", "Google Chrome Helper (Renderer)", "x --type=renderer --extension-process", "Extension"},
		{"Google Chrome", "Google Chrome Helper (GPU)", "x --type=gpu-process", "GPU Process"},
		{"Google Chrome", "Google Chrome Helper", "x --type=utility --utility-sub-type=network.mojom.NetworkService", "Network Service"},
		{"Visual Studio Code", "Code Helper (Plugin)", "x --type=utility --utility-sub-type=node.mojom.NodeService", "Node Service"},
		{"Visual Studio Code", "Code Helper (Plugin)", "x --type=utility --utility-sub-type=other.mojom.Thing", "Plugin"},
		{"Loom", "Loom Helper (Renderer)", "", "Renderer"},
		{"Loom", "Loom Helper", "", "Loom Helper"},
		{"Docker", "com.docker.backend", "", "com.docker.backend"},
		{"Slack", "Slack Helper", "x --type=utility", "Utility"},
	}
	for _, c := range cases {
		if got := roleLabel(c.name, c.cmdline); got != c.want {
			t.Errorf("%s / %s / %s: want %q, got %q", c.app, c.name, c.cmdline, c.want, got)
		}
	}
}
