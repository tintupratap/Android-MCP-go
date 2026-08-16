package scrcpy

import (
	"fmt"
	"os"
	"strings"

	"github.com/tintupratap/Android-MCP-go/internal/config"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
)

type ScrcpyProfile struct {
	Name         string
	Args         []string
	VideoCodec   string
	VideoEncoder string
	Renderer     string
	Audio        string
	Bitrate      string
	Optimized    bool
	FallbackStep int
}

func BuildArgs(userPrefs *config.ScrcpyPreferences, serial string, title string) []string {
	caps := SystemCapabilities{
		Binary: BinaryCapabilities{
			SupportsRenderDriver: true,
			SupportsVideoCodec:   true,
			SupportsAudioSource:  true,
			SupportedFlags: map[string]bool{
				"--stay-awake":     true,
				"--always-on-top": true,
			},
		},
	}
	prof := ResolveOptimalProfile(userPrefs, caps, serial, title)
	return prof.Args
}

func ResolveOptimalProfile(userPrefs *config.ScrcpyPreferences, caps SystemCapabilities, serial string, title string) ScrcpyProfile {
	if title == "" {
		title = fmt.Sprintf("Android-MCP — %s", serial)
	}

	profileName := "auto"
	if userPrefs != nil && userPrefs.Profile != "" && userPrefs.Profile != "auto" {
		profileName = userPrefs.Profile
	}

	baseArgs := []string{"--window-title", title}
	if userPrefs == nil || !userPrefs.DisableAlwaysOnTop {
		baseArgs = append(baseArgs, "--always-on-top")
	}
	if serial != "" {
		baseArgs = append(baseArgs, "-s", serial)
	}

	// 1. Determine Video Codec
	videoCodec := "h265"
	if userPrefs != nil && userPrefs.VideoCodec != "" && userPrefs.VideoCodec != "auto" {
		videoCodec = userPrefs.VideoCodec
	} else if profileName == "compatibility" {
		videoCodec = "h264"
	}

	// 2. Determine Video Bitrate based on WiFi vs USB
	bitrate := "4M"
	if userPrefs != nil && userPrefs.VideoBitrate != "" && userPrefs.VideoBitrate != "auto" {
		bitrate = userPrefs.VideoBitrate
	} else if caps.Device.IsWiFi {
		bitrate = "4M" // Low-latency conscious WiFi bitrate
	} else {
		bitrate = "8M" // High bandwidth USB bitrate
	}

	// 3. Determine Renderer
	renderer := caps.Host.NativeRenderer
	if userPrefs != nil && userPrefs.RenderDriver != "" && userPrefs.RenderDriver != "auto" {
		renderer = userPrefs.RenderDriver
	} else if profileName == "compatibility" {
		renderer = ""
	}

	// 4. Build argument list dynamically checking binary capabilities
	args := append([]string{}, baseArgs...)

	if caps.Binary.SupportsVideoCodec && videoCodec != "" {
		args = append(args, "--video-codec", videoCodec)
	}
	if videoBitrateArg(userPrefs) != "" {
		args = append(args, "--video-bit-rate", bitrate)
	}

	if caps.Binary.SupportsRenderDriver && renderer != "" {
		args = append(args, "--render-driver", renderer)
	}

	if userPrefs != nil && userPrefs.DisplayID != "" && caps.Binary.SupportsDisplayID {
		args = append(args, "--display-id", userPrefs.DisplayID)
	}

	audioState := "enabled"
	if userPrefs != nil && (userPrefs.Audio == "disabled" || userPrefs.AudioSource == "disabled") {
		audioState = "disabled"
		if caps.Binary.SupportedFlags["--no-audio"] {
			args = append(args, "--no-audio")
		}
	} else if caps.Binary.SupportsAudioSource {
		args = append(args, "--audio-source", "playback")
	}

	if userPrefs != nil && userPrefs.StayAwake && caps.Binary.SupportedFlags["--stay-awake"] {
		args = append(args, "--stay-awake")
	}

	if extraArgs := os.Getenv("ANDROID_MCP_SCRCPY_ARGS"); extraArgs != "" {
		args = append(args, strings.Fields(extraArgs)...)
	}

	return ScrcpyProfile{
		Name:         profileName,
		Args:         args,
		VideoCodec:   videoCodec,
		Renderer:     renderer,
		Audio:        audioState,
		Bitrate:      bitrate,
		Optimized:    true,
		FallbackStep: 0,
	}
}

func videoBitrateArg(prefs *config.ScrcpyPreferences) string {
	return "--video-bit-rate"
}

func DegradeProfile(current ScrcpyProfile, caps SystemCapabilities, lastErr error) (ScrcpyProfile, bool) {
	next := current
	next.FallbackStep++

	logging.Warnf("scrcpy profile degradation triggered (step %d, reason: %v)", next.FallbackStep, lastErr)

	switch next.FallbackStep {
	case 1:
		// Step 1: Remove render-driver optimization if unsupported or failed
		if next.Renderer != "" {
			next.Renderer = ""
			next.Name = "reduced_optimized"
			next.Args = filterOutFlag(next.Args, "--render-driver")
			return next, true
		}
		fallthrough
	case 2:
		// Step 2: Fallback from H.265 to standard H.264
		if next.VideoCodec == "h265" {
			next.VideoCodec = "h264"
			next.Name = "h264_fallback"
			next.Args = replaceFlagValue(next.Args, "--video-codec", "h264")
			return next, true
		}
		fallthrough
	case 3:
		// Step 3: Disable audio if audio pipeline fails
		if next.Audio != "disabled" {
			next.Audio = "disabled"
			next.Name = "no_audio_fallback"
			next.Args = filterOutFlag(next.Args, "--audio-source")
			next.Args = filterOutFlag(next.Args, "--audio-codec")
			if caps.Binary.SupportedFlags["--no-audio"] {
				next.Args = append(next.Args, "--no-audio")
			}
			return next, true
		}
		fallthrough
	case 4:
		// Step 4: Minimal safe profile
		next.Name = "minimal_safe"
		next.Optimized = false
		title := "Android-MCP"
		serial := ""
		for i, a := range current.Args {
			if a == "--window-title" && i+1 < len(current.Args) {
				title = current.Args[i+1]
			}
			if a == "-s" && i+1 < len(current.Args) {
				serial = current.Args[i+1]
			}
		}
		minArgs := []string{"--window-title", title}
		if serial != "" {
			minArgs = append(minArgs, "-s", serial)
		}
		next.Args = minArgs
		return next, true
	default:
		// Stop fallback attempts
		return current, false
	}
}

func filterOutFlag(args []string, flagName string) []string {
	var result []string
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == flagName {
			skipNext = true
			continue
		}
		if strings.HasPrefix(arg, flagName+"=") {
			continue
		}
		result = append(result, arg)
	}
	return result
}

func replaceFlagValue(args []string, flagName, val string) []string {
	var result []string
	skipNext := false
	replaced := false
	for i := 0; i < len(args); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		if args[i] == flagName && i+1 < len(args) {
			result = append(result, flagName, val)
			skipNext = true
			replaced = true
			continue
		}
		if strings.HasPrefix(args[i], flagName+"=") {
			result = append(result, fmt.Sprintf("%s=%s", flagName, val))
			replaced = true
			continue
		}
		result = append(result, args[i])
	}
	if !replaced {
		result = append(result, flagName, val)
	}
	return result
}
