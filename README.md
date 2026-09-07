# doman

`doman` is a command-line utility for various simple tasks.

Originally built as a collection of tools for doing work, this program has evolved into a collection of general utilities and an opportunity to try new things in Go.

## Installation

### Quick Install

```bash
go install doman.sh/doman/doman@latest
```

### Unix/macOS/Linux

```bash
curl -fsSL https://doman.sh/doman/install.sh | bash
```

Or download and run manually:

```bash
curl -fsSL https://doman.sh/doman/install.sh -o install.sh
bash install.sh
```

### Windows (PowerShell)

```powershell
irm https://doman.sh/doman/install.ps1 | iex
```

Or download and run manually:

```powershell
Invoke-WebRequest -Uri https://doman.sh/doman/install.ps1 -OutFile install.ps1
.\install.ps1
```

**Note:** If you encounter execution policy restrictions, run:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## Favorite Commands

These are the commands I use most often:

1. `ask` – Get a question answered by AI (requires your own OpenAI API key)
2. `ask live` – A full screen, back and forth chat. Resume it with `-c`, and type `/quick` anywhere in a message for a fast answer from the smaller model.
3. `ip` – Display your public and private IP address.
4. `git author` – Display the name & email of the current Git user.
