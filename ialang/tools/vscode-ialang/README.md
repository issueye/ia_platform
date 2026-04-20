# ialang VS Code Extension

This extension adds:

- Syntax highlighting for `.ia` files
- Language Server Protocol support:
  - diagnostics (delimiter mismatch + possible missing semicolon warnings)
  - keyword/builtin completion
  - hover docs for core keywords
  - go to definition (same-file `let` / `function` / `class`)

## Development

1. Open this folder in VS Code:
   - `tools/vscode-ialang`
2. Install dependencies:
   - `npm install`
3. Compile:
   - `npm run compile`
4. Press `F5` to launch the Extension Development Host.
5. Open any `.ia` file to activate the extension.

Native module completion/hover data is generated from:
- `docs/2026-04-07/NATIVE_MODULES.md`

Generate manually:

```bash
npm run gen:native
```

## Package

```bash
npm install -g @vscode/vsce
npm run compile
npm run package
```

This produces a `.vsix` package for installation.

## Package Script (Repo Root)

From repository root:

```powershell
.\scripts\package-vscode-ialang.ps1
```

Optional:

```powershell
.\scripts\package-vscode-ialang.ps1 -InstallDeps -OutFile dist\ialang.vsix
```
