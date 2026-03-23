package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"flutter-deploy/internal/builder"
	"flutter-deploy/internal/config"
	"flutter-deploy/internal/diawi"
	"flutter-deploy/internal/firebase"
	"flutter-deploy/internal/telegram"
	"flutter-deploy/internal/version"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen, color.Bold).SprintFunc()
	red    = color.New(color.FgRed, color.Bold).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
)

func Execute() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n %s %s\n", red("✗"), err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Println()
	fmt.Println(cyan("╔══════════════════════════════════╗"))
	fmt.Println(cyan("║    🚀 Flutter Deploy CLI         ║"))
	fmt.Println(cyan("╚══════════════════════════════════╝"))
	fmt.Println()

	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	fmt.Printf(" %s Config loaded\n", green("✓"))

	// 2. Select platform
	var platform string
	platformPrompt := &survey.Select{
		Message: "Select platform:",
		Options: []string{"Android", "iOS"},
	}
	if err := survey.AskOne(platformPrompt, &platform); err != nil {
		return err
	}
	fmt.Printf(" %s Platform: %s\n", green("✓"), yellow(platform))

	// 3. Select flavor
	var flavor string
	prompt := &survey.Select{
		Message: "Select flavor:",
		Options: []string{"develop", "production"},
	}
	if err := survey.AskOne(prompt, &flavor); err != nil {
		return err
	}
	fmt.Printf(" %s Flavor: %s\n", green("✓"), yellow(flavor))

	// 3. Select groups
	fbKey := "dev"
	if flavor == "production" {
		fbKey = "production"
	}
	fbCfg, ok := cfg.Firebase[fbKey]
	if !ok {
		return fmt.Errorf("firebase config not found for %q", fbKey)
	}

	var selectedGroups string
	if fbCfg.Groups != "" {
		allGroups := strings.Split(fbCfg.Groups, ",")
		for i := range allGroups {
			allGroups[i] = strings.TrimSpace(allGroups[i])
		}
		var chosen []string
		groupPrompt := &survey.MultiSelect{
			Message: "Select groups (optional):",
			Options: allGroups,
		}
		if err := survey.AskOne(groupPrompt, &chosen); err != nil {
			return err
		}
		selectedGroups = strings.Join(chosen, ",")
		if selectedGroups != "" {
			fmt.Printf(" %s Groups: %s\n", green("✓"), yellow(selectedGroups))
		} else {
			fmt.Printf(" %s Groups: none\n", yellow("⚠"))
		}
	}
	fbCfg.Groups = selectedGroups
	fmt.Println()

	// 4. Bump version
	newVersion, err := version.Bump()
	if err != nil {
		return err
	}
	fmt.Printf(" %s Version bumped → %s\n", green("✓"), yellow(newVersion))

	// 5. Build
	var result *builder.Result
	if platform == "Android" {
		fmt.Printf("\n %s Building APK [%s]...\n", cyan("⟳"), flavor)
		result, err = builder.BuildAPK(flavor)
	} else {
		fmt.Printf("\n %s Building IPA [%s]...\n", cyan("⟳"), flavor)
		result, err = builder.BuildIPA(flavor)
	}
	if err != nil {
		return err
	}
	fmt.Printf(" %s Built: %s\n", green("✓"), result.FileName)

	// 6. Get branch name and extract ticket ID
	branchOut, err := exec.Command("git", "-C", "..", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("get git branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOut))
	ticketID := branch
	if idx := strings.Index(branch, "_"); idx != -1 {
		ticketID = branch[:idx]
	}
	fmt.Printf(" %s Branch: %s → Ticket: %s\n", green("✓"), yellow(branch), yellow(ticketID))

	// 7. Upload
	envLabel := "DEV"
	if flavor == "production" {
		envLabel = "PROD"
	}
	releaseNotes := fmt.Sprintf("[%s] %s", ticketID, branch)
	var downloadLink string

	if platform == "Android" {
		// Upload to Firebase
		fmt.Printf("\n %s Uploading to Firebase [%s]...\n", cyan("⟳"), fbKey)
		fbResult, err := firebase.Upload(&fbCfg, result.FilePath, releaseNotes)
		if err != nil {
			return err
		}
		downloadLink = fbResult.ConsoleLink
		fmt.Printf(" %s Firebase upload complete!\n", green("✓"))
	} else {
		// Upload to Diawi
		fmt.Printf("\n %s Uploading to Diawi...\n", cyan("⟳"))
		link, err := diawi.Upload(cfg.Diawi.Token, result.FilePath)
		if err != nil {
			return err
		}
		downloadLink = link
		fmt.Printf(" %s Diawi upload complete!\n", green("✓"))
	}

	// 8. Send Telegram notification
	platformLabel := "Android"
	if platform == "iOS" {
		platformLabel = "iOS"
	}
	teleMsg := fmt.Sprintf("🚀 HCN SX %s %s [%s]\n\n📋 %s\n📦 Version: %s\n\n🔗 %s",
		platformLabel, envLabel, ticketID, branch, newVersion, downloadLink)

	fmt.Printf("\n %s Sending Telegram notification...\n", cyan("⟳"))
	if err := telegram.SendMessage(&cfg.Telegram, teleMsg); err != nil {
		return err
	}
	fmt.Printf(" %s Telegram sent!\n", green("✓"))

	// Done
	fmt.Println()
	fmt.Println(green("══════════════════════════════════"))
	fmt.Printf(green(" ✅ Deploy %s [%s] %s done!\n"), platformLabel, flavor, newVersion)
	fmt.Println(green("══════════════════════════════════"))

	return nil
}
