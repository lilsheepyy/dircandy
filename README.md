
# 📦 dircandy – Terminal File Manager

A stylish terminal UI (TUI) for managing files with `cp`, `mv`, and `rm -rf` functionality, written in Go.

## 🗂 Features

- Navigate directories with intuitive arrow keys
- Multi-select files and directories for batch operations
- Copy, move, or delete with ease
- Start in the current working directory
- Clean UI for a smooth experience

## 🚀 Installation

### Build and Install
```bash
git clone https://github.com/yourusername/dircandy.git
cd dircandy
make install
```

By default, installs to `/opt/dircandy`. To use it from anywhere:
```bash
export PATH=$PATH:/opt/dircandy
```
Or set an alias:
```bash
alias dircandy='/opt/dircandy/dircandy'
```

### Custom Install Directory
```bash
sudo make install INSTALL_DIR=/usr/local/bin
```

## 🏃 Usage
```bash
dircandy
```
- Choose Copy, Move, or Remove with keys or arrows
- Navigate directories: ↑/↓ to move, →/Enter to enter dir, ←/Backspace to go up
- Select files: Space
- Confirm actions: Tab or Enter
- Quit: q

## 🧹 Clean Build Files
```bash
make clean
```

## 📄 License
MIT License
