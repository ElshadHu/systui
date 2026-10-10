package apps

import "strings"

var typeRoles = map[string]string{
	"renderer":         "Renderer",
	"gpu-process":      "GPU Process",
	"zygote":           "Zygote",
	"broker":           "Broker",
	"ppapi":            "Plugin",
	"crashpad-handler": "Crash Handler",
	"utility":          "Utility",
}

var utilityRoles = map[string]string{
	"network.mojom.NetworkService":  "Network Service",
	"storage.mojom.StorageService":  "Storage Service",
	"audio.mojom.AudioService":      "Audio Service",
	"video_capture.mojom.":          "Video Capture",
	"node.mojom.NodeService":        "Node Service",
	"data_decoder.mojom.":           "Data Decoder",
	"printing.mojom.":               "Printing",
	"unzip.mojom.":                  "Unzip",
	"chrome.mojom.FileUtilService":  "File Util",
	"chrome.mojom.ProcessorMetrics": "Metrics",
}

// roleLabel names a helper process from its command line flags
func roleLabel(procName, cmdline string) string {
	if strings.Contains(cmdline, "--extension-process") {
		return "Extension"
	}
	helper := helperSuffix(procName)
	kind := flagValue(cmdline, "--type=")
	if kind == "utility" {
		if sub := flagValue(cmdline, "--utility-sub-type="); sub != "" {
			for prefix, role := range utilityRoles {
				if strings.HasPrefix(sub, prefix) {
					return role
				}
			}
		}
		if helper != "" {
			return helper
		}
	}
	if role, ok := typeRoles[kind]; ok {
		return role
	}
	if helper != "" {
		return helper
	}
	return procName
}

func helperSuffix(procName string) string {
	_, rest, ok := strings.Cut(procName, "Helper")
	if !ok {
		return ""
	}
	rest = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(rest), "("), ")")
	return strings.TrimSpace(rest)
}

func flagValue(cmdline, flag string) string {
	i := strings.Index(cmdline, flag)
	if i < 0 {
		return ""
	}
	value := cmdline[i+len(flag):]
	if end := strings.IndexAny(value, " \t"); end >= 0 {
		value = value[:end]
	}
	return value
}
