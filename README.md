# agef

> High-performance, native Linux package for **mass folder encryption and decryption** using [age](https://github.com/FiloSottile/age).
It can encrypy/decrypt all the files of an folder separately given the folder. Each file gets encrypted via the [age](https://github.com/FiloSottile/age) tool. The main reason this was created
was to allow users to mass encrypt files.

[![CI](https://github.com/your-username/agef/actions/workflows/ci.yml/badge.svg)](https://github.com/your-username/agef/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release Packages](https://img.shields.io/badge/packages-RPM%20%7C%20DEB%20%7C%20Tarball-blue.svg)](https://github.com/your-username/agef/releases)
[![Go 1.22+](https://img.shields.io/badge/go-1.22+-00ADD8.svg)](https://go.dev/)

---

## ⚡ Why `agef`?

Rather than shelling out to external scripts, `agef` is built directly on top of the official **`filippo.io/age` Go library** created by Filippo Valsorda (former cryptography lead for Go at Google).

- **It's Fast**: Uses native in-memory streaming cryptography and parallel goroutine worker pools. Encrypts and decrypts hundreds of files concurrently without process-spawning overhead.
- **Zero External Dependencies**: The complete X25519 & ChaCha20-Poly1305 engine is compiled directly into a single statically-linked binary.
- **In-Place Output**: Encrypted (`<folder>_encrypted`) and decrypted (`<folder>_decrypted`) directories are created right in the same location where the original was.
- **Automatic Config**: Automatically creates and locks `~/.config/agef/` (`chmod 700`) on first launch to manage keys securely.

---

## Installation Guide for All Linux Distros

### 1. Fedora / RHEL / CentOS (`dnf`)
Install directly via DNF from the latest GitHub Release:

```bash
sudo dnf install https://github.com/Lazy-Pir8/agef/releases/latest/download/agef.rpm
```

*Or via Fedora Copr:*
```bash
sudo dnf copr enable Lazy-Pir8/agef
sudo dnf install agef
```

---

### 2. Ubuntu / Debian / Linux Mint (`apt`)
Install the native `.deb` package directly:

```bash
curl -sLO https://github.com/Lazy-Pir8/agef/releases/latest/download/agef.deb
sudo apt install ./agef.deb
rm agef.deb
```

---

### 3. Arch Linux / Manjaro (`pacman`)

#### Option A: Quick Standalone Binary
```bash
curl -sSL https://github.com/Lazy-Pir8/agef/releases/latest/download/agef-linux-amd64.tar.gz | sudo tar -xz -C /usr/local/bin
```

#### Option B: Build with `PKGBUILD`
```bash
git clone https://github.com/Lazy-Pir8/agef.git
cd agef/packaging
makepkg -si
```

---

### 4. Build from Source
```bash
git clone https://github.com/Lazy-Pir8/agef.git
cd agef
make install
```

---

## 🚀 Usage

Once installed, simply type **`agef`** in any terminal:

```bash
agef
```

You get a interactive terminal interface:

```text
╔══════════════════════════════════════════════════════╗
║                         AGEF                       ║
╚══════════════════════════════════════════════════════╝
  [1] Initialise Keys  (Generate keypair in ~/.config/agef)
  [2] Encrypt Folder   (Create <folder>_encrypted)
  [3] Decrypt Folder   (Create <folder>_decrypted)
  [4] View Key Info    (Check active keys & status)
  [5] Exit
────────────────────────────────────────────────────────
Select an option [1-5]:
```

---

### The 3 Core Operations

#### (1) Initialise Keys
Select **Option 1**:
- Generates a native X25519 asymmetric keypair.
- Automatically stores keys in your Linux user config directory:
  - `~/.config/agef/key.txt` (Private identity key, locked with `chmod 600`)
  - `~/.config/agef/key.pub` (Public recipient key, `chmod 644`)

#### (2) Encrypt Folder
Select **Option 2**:
- Prompts: `Enter path to the folder you want to encrypt: /path/to/my_data`
- Creates `/path/to/my_data_encrypted` in the same location.
- Concurrently encrypts every file separately into a `.age` file.
- Replicates directory hierarchy, empty folders, and file timestamps.
- **Leaves The original folder untouched.**

#### (3) Decrypt Folder
Select **Option 3**:
- Prompts: `Enter path to the encrypted folder: /path/to/my_data_encrypted`
- Creates `/path/to/my_data_decrypted` in the same location.
- Decrypts all files back to their original names and formats.
- **Leaves the encrypted folder untouched.**

---

##  Direct CLI Commands (Automation & CI/CD)

Bypass interactive menus by passing arguments directly:

```bash
# Initialise keys
agef init

# Encrypt folder
agef encrypt /path/to/documents
# Output: /path/to/documents_encrypted

# Decrypt folder
agef decrypt /path/to/documents_encrypted
# Output: /path/to/documents_decrypted

# Check key status
agef status
```

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
