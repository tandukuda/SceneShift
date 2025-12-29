# Installation

SceneShift is built for **Windows 10 and 11**.

## Method 1: Scoop (Recommended)
The easiest way to install and keep SceneShift updated is using the [Scoop](https://scoop.sh/) package manager.

Open PowerShell and run:

```powershell
# 1. Add the bucket
scoop bucket add tandukuda https://github.com/tandukuda/scoop-bucket

# 2. Install SceneShift
scoop install sceneshift
```

## Method 2: Manual Installation
If you prefer not to use a package manager, you can run the standalone binary.

1.  **Download:** Get the latest release from the [Releases Page](https://github.com/tandukuda/SceneShift/releases).
2.  **Organize:** Create a new folder (e.g., `Documents/SceneShift`) and move `SceneShift.exe` inside.
    !!! warning "Important"
        Do not run the app directly from your Desktop or Downloads! On the first run, SceneShift automatically generates `config.yaml` and `theme.yaml`. Keeping it in a separate folder ensures your files stay organized.
3.  **Run:** Double-click `SceneShift.exe`.
    * *Note: Windows will request Administrator access to manage processes.*
4.  **Shortcut:** Right-click `SceneShift.exe` → **Send to** → **Desktop (create shortcut)** for quick access later.


## Method 3: Build From Source
If you prefer to build it yourself, you will need **Go 1.21+** installed.

```bash
# 1. Clone the repo
git clone https://github.com/tandukuda/SceneShift.git
cd SceneShift

# 2. Install resource tool (for icon embedding)
go install github.com/akavel/rsrc@latest

# 3. Build with assets
rsrc -manifest sceneshift.manifest -ico assets/icon.ico -o sceneshift.syso
go build -ldflags "-s -w" -o SceneShift.exe
```
