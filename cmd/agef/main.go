package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/agef/internal/config"
	"github.com/user/agef/internal/crypto"
	"github.com/user/agef/internal/keys"
	"github.com/user/agef/internal/ui"
)

const version = "1.0.0"

func main() {
	// Automatically ensure ~/.config/agef exists
	_, err := config.EnsureConfigDir()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to initialize config directory: %v", err))
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init", "keygen":
			runInit()
		case "encrypt", "enc":
			folder := ""
			if len(os.Args) > 2 {
				folder = os.Args[2]
			}
			runEncrypt(folder)
		case "decrypt", "dec":
			folder := ""
			if len(os.Args) > 2 {
				folder = os.Args[2]
			}
			runDecrypt(folder)
		case "status", "info":
			runStatus()
		case "version", "-v", "--version":
			fmt.Printf("agef version %s\n", version)
		case "help", "-h", "--help":
			printUsage()
		default:
			ui.Error(fmt.Sprintf("Unknown command: %s", os.Args[1]))
			printUsage()
			os.Exit(1)
		}
		return
	}

	interactiveMenu()
}

func printUsage() {
	fmt.Println("Usage: agef [COMMAND] [ARGS]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  (none)              Launch interactive terminal menu")
	fmt.Println("  init                Initialise public and private keypair in ~/.config/agef")
	fmt.Println("  encrypt [DIR]       Encrypt folder into <DIR>_encrypted in same location")
	fmt.Println("  decrypt [DIR]       Decrypt folder into <DIR>_decrypted in same location")
	fmt.Println("  status              Show key details and storage paths")
	fmt.Println("  version             Display version information")
	fmt.Println("  help                Show this help message")
}

func interactiveMenu() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("\n%s%s╔══════════════════════════════════════════════════════╗%s\n", ui.Bold, ui.Cyan, ui.Reset)
		fmt.Printf("%s%s║                      🛡️   AGEF                       ║%s\n", ui.Bold, ui.Cyan, ui.Reset)
		fmt.Printf("%s%s╚══════════════════════════════════════════════════════╝%s\n", ui.Bold, ui.Cyan, ui.Reset)
		fmt.Printf("  %s[1]%s Initialise Keys  %s(Generate keypair in ~/.config/agef)%s\n", ui.Bold, ui.Reset, ui.Dim, ui.Reset)
		fmt.Printf("  %s[2]%s Encrypt Folder   %s(Create <folder>_encrypted in same place)%s\n", ui.Bold, ui.Reset, ui.Dim, ui.Reset)
		fmt.Printf("  %s[3]%s Decrypt Folder   %s(Create <folder>_decrypted in same place)%s\n", ui.Bold, ui.Reset, ui.Dim, ui.Reset)
		fmt.Printf("  %s[4]%s View Key Info    %s(Check active keys & status)%s\n", ui.Bold, ui.Reset, ui.Dim, ui.Reset)
		fmt.Printf("  %s[5]%s Exit\n", ui.Bold, ui.Reset)
		fmt.Printf("%s────────────────────────────────────────────────────────%s\n", ui.Cyan, ui.Reset)
		fmt.Printf("Select an option [1-5]: ")

		if !scanner.Scan() {
			break
		}
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			runInit()
		case "2":
			runEncrypt("")
		case "3":
			runDecrypt("")
		case "4":
			runStatus()
		case "5", "q", "Q", "exit":
			fmt.Printf("\nGoodbye! 👋\n\n")
			return
		default:
			ui.Error("Invalid selection. Choose an option between 1 and 5.")
		}

		fmt.Printf("\nPress Enter to continue...")
		scanner.Scan()
	}
}

func runInit() {
	ui.Header("🔑 INITIALISE AGEF ENCRYPTION KEYS")

	configDir, err := config.DefaultConfigDir()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to resolve config dir: %v", err))
		return
	}

	fmt.Printf("Keys will be stored in standard Linux location:\n  %s%s%s\n\n", ui.Cyan, configDir, ui.Reset)

	if config.HasStandardKeys() {
		ui.Warn(fmt.Sprintf("Existing keys found in: %s", configDir))
		confirm := ui.Prompt("Overwrite existing keys? (WARNING: Old encryptions will become undecryptable!) [y/N]", "N")
		if !strings.EqualFold(confirm, "y") && !strings.EqualFold(confirm, "yes") {
			fmt.Println("Aborted key generation. Keeping existing keys.")
			return
		}
	}

	fmt.Println("Generating cryptographic X25519 keypair...")
	privPath, pubPath, pubKeyStr, err := keys.GenerateKeypair(configDir)
	if err != nil {
		ui.Error(fmt.Sprintf("Key generation failed: %v", err))
		return
	}

	ui.Success("Keys initialised successfully!")
	fmt.Printf("  • %sStorage Path:%s  %s%s%s\n", ui.Bold, ui.Reset, ui.Cyan, configDir, ui.Reset)
	fmt.Printf("  • %sPrivate Key:%s   %s%s%s %s(KEEP SECRET, NEVER SHARE!)%s\n", ui.Bold, ui.Reset, ui.Red, privPath, ui.Reset, ui.Dim, ui.Reset)
	fmt.Printf("  • %sPublic Key:%s    %s%s%s %s(Safe to share)%s\n\n", ui.Bold, ui.Reset, ui.Green, pubPath, ui.Reset, ui.Dim, ui.Reset)
	fmt.Printf("Recipient Public Key String:\n%s%s%s%s\n", ui.Bold, ui.Cyan, pubKeyStr, ui.Reset)
}

func runEncrypt(folder string) {
	ui.Header("🔒 ENCRYPT FOLDER (MASS ENCRYPTION)")

	for folder == "" {
		folder = ui.Prompt("Enter path to the folder you want to encrypt", "")
		if folder == "" {
			ui.Error("Folder path cannot be empty.")
		}
	}

	absSrc, err := filepath.Abs(folder)
	if err != nil {
		ui.Error(fmt.Sprintf("Invalid path: %v", err))
		return
	}

	info, err := os.Stat(absSrc)
	if err != nil || !info.IsDir() {
		ui.Error(fmt.Sprintf("Directory does not exist: %s", absSrc))
		return
	}

	// In the same place where original folder was
	parentDir := filepath.Dir(absSrc)
	baseName := filepath.Base(absSrc)
	destDir := filepath.Join(parentDir, baseName+"_encrypted")

	// Ensure keys exist
	if !config.HasStandardKeys() {
		ui.Warn("No encryption keys found. Initialising keys first...")
		runInit()
		if !config.HasStandardKeys() {
			ui.Error("Cannot proceed without encryption keys.")
			return
		}
	}

	recipient, err := keys.LoadRecipient("")
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load recipient key: %v", err))
		return
	}

	fmt.Printf("  • %sSource Folder:%s      %s%s%s\n", ui.Bold, ui.Reset, ui.Cyan, absSrc, ui.Reset)
	fmt.Printf("  • %sEncrypted Folder:%s   %s%s%s\n\n", ui.Bold, ui.Reset, ui.Green, destDir, ui.Reset)

	if _, err := os.Stat(destDir); err == nil {
		ui.Warn(fmt.Sprintf("Encrypted folder already exists at: %s", destDir))
		confirm := ui.Prompt("Overwrite/update files in destination? [y/N]", "N")
		if !strings.EqualFold(confirm, "y") && !strings.EqualFold(confirm, "yes") {
			fmt.Println("Encryption cancelled.")
			return
		}
	}

	fmt.Println("Encrypting files recursively...")
	progress := func(completed int64, total int64, currentPath string) {
		shortPath := currentPath
		if len(shortPath) > 50 {
			shortPath = "..." + shortPath[len(shortPath)-47:]
		}
		fmt.Printf("\r\033[K[%d/%d] Encrypting: %s", completed, total, shortPath)
	}

	count, err := crypto.EncryptFolder(absSrc, destDir, recipient, progress)
	fmt.Print("\r\033[K")
	if err != nil {
		ui.Error(fmt.Sprintf("Encryption encountered an error: %v", err))
		return
	}

	ui.Success("Mass Encryption Complete!")
	fmt.Printf("  • %sOriginal Folder (untouched):%s %s\n", ui.Bold, ui.Reset, absSrc)
	fmt.Printf("  • %sNew Encrypted Folder:%s        %s%s%s\n", ui.Bold, ui.Reset, ui.Green, destDir, ui.Reset)
	fmt.Printf("  • %sFiles Encrypted:%s             %d\n", ui.Bold, ui.Reset, count)
	fmt.Printf("\n%s💡 You can now safely backup, move, or push '%s_encrypted' anywhere.%s\n", ui.Cyan, baseName, ui.Reset)
}

func runDecrypt(folder string) {
	ui.Header("🔓 DECRYPT FOLDER")

	for folder == "" {
		folder = ui.Prompt("Enter path to the encrypted folder", "")
		if folder == "" {
			ui.Error("Folder path cannot be empty.")
		}
	}

	absSrc, err := filepath.Abs(folder)
	if err != nil {
		ui.Error(fmt.Sprintf("Invalid path: %v", err))
		return
	}

	info, err := os.Stat(absSrc)
	if err != nil || !info.IsDir() {
		ui.Error(fmt.Sprintf("Directory does not exist: %s", absSrc))
		return
	}

	parentDir := filepath.Dir(absSrc)
	baseName := filepath.Base(absSrc)
	var destDir string
	if strings.HasSuffix(baseName, "_encrypted") {
		destDir = filepath.Join(parentDir, strings.TrimSuffix(baseName, "_encrypted")+"_decrypted")
	} else {
		destDir = filepath.Join(parentDir, baseName+"_decrypted")
	}

	identity, err := keys.LoadIdentity("")
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load private identity key: %v", err))
		return
	}

	fmt.Printf("  • %sEncrypted Source:%s   %s%s%s\n", ui.Bold, ui.Reset, ui.Cyan, absSrc, ui.Reset)
	fmt.Printf("  • %sDecrypted Folder:%s   %s%s%s\n\n", ui.Bold, ui.Reset, ui.Green, destDir, ui.Reset)

	if _, err := os.Stat(destDir); err == nil {
		ui.Warn(fmt.Sprintf("Decrypted destination folder already exists at: %s", destDir))
		confirm := ui.Prompt("Overwrite/update files in destination? [y/N]", "N")
		if !strings.EqualFold(confirm, "y") && !strings.EqualFold(confirm, "yes") {
			fmt.Println("Decryption cancelled.")
			return
		}
	}

	fmt.Println("Decrypting files recursively...")
	progress := func(completed int64, total int64, currentPath string) {
		shortPath := currentPath
		if len(shortPath) > 50 {
			shortPath = "..." + shortPath[len(shortPath)-47:]
		}
		fmt.Printf("\r\033[K[%d/%d] Decrypting: %s", completed, total, shortPath)
	}

	count, err := crypto.DecryptFolder(absSrc, destDir, identity, progress)
	fmt.Print("\r\033[K")
	if err != nil {
		ui.Error(fmt.Sprintf("Decryption error: %v", err))
		return
	}

	ui.Success("Decryption Complete!")
	fmt.Printf("  • %sEncrypted Source (untouched):%s %s\n", ui.Bold, ui.Reset, absSrc)
	fmt.Printf("  • %sNew Decrypted Folder:%s        %s%s%s\n", ui.Bold, ui.Reset, ui.Green, destDir, ui.Reset)
	fmt.Printf("  • %sFiles Decrypted:%s             %d\n", ui.Bold, ui.Reset, count)
}

func runStatus() {
	ui.Header("ℹ️  AGEF STATUS & KEY INFO")

	configDir, _ := config.DefaultConfigDir()
	privPath, _ := config.PrivateKeyPath()
	pubPath, _ := config.PublicKeyPath()

	fmt.Printf("  • %sStandard Config Directory:%s %s%s%s\n", ui.Bold, ui.Reset, ui.Cyan, configDir, ui.Reset)

	if _, err := os.Stat(privPath); err == nil {
		fmt.Printf("  • %sPrivate Key:%s               %sPresent%s (%s)\n", ui.Bold, ui.Reset, ui.Green, ui.Reset, privPath)
	} else {
		fmt.Printf("  • %sPrivate Key:%s               %sNot Found%s\n", ui.Bold, ui.Reset, ui.Red, ui.Reset)
	}

	if _, err := os.Stat(pubPath); err == nil {
		fmt.Printf("  • %sPublic Key:%s                %sPresent%s (%s)\n", ui.Bold, ui.Reset, ui.Green, ui.Reset, pubPath)
		pubStr, err := keys.ReadPublicKeyString(pubPath)
		if err == nil {
			fmt.Printf("  • %sRecipient String:%s          %s%s%s\n", ui.Bold, ui.Reset, ui.Cyan, pubStr, ui.Reset)
		}
	} else {
		fmt.Printf("  • %sPublic Key:%s                %sNot Found%s\n", ui.Bold, ui.Reset, ui.Red, ui.Reset)
	}

	fmt.Println()
}
