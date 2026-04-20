import {
  CompletionItem,
  CompletionItemKind,
  createConnection,
  Diagnostic,
  DiagnosticSeverity,
  Hover,
  InitializeParams,
  InitializeResult,
  Location,
  Position,
  ProposedFeatures,
  TextDocumentSyncKind,
  TextDocumentPositionParams,
  TextDocuments
} from "vscode-languageserver/node";
import { TextDocument } from "vscode-languageserver-textdocument";
import { nativeSymbols } from "./generated/nativeSymbols";

const connection = createConnection(ProposedFeatures.all);
const documents: TextDocuments<TextDocument> = new TextDocuments(TextDocument);

const keywords = [
  "import",
  "export",
  "from",
  "class",
  "new",
  "this",
  "super",
  "extends",
  "let",
  "await",
  "async",
  "function",
  "return",
  "throw",
  "if",
  "else",
  "while",
  "for",
  "break",
  "continue",
  "try",
  "catch",
  "finally",
  "true",
  "false"
];

const keywordDocs: Record<string, string> = {
  import: "Import named symbols from a module.",
  export: "Export declaration from current module.",
  class: "Define a class.",
  extends: "Set class inheritance.",
  function: "Define a function.",
  async: "Define async function or method.",
  await: "Wait for an awaitable/promise value.",
  let: "Declare mutable binding.",
  this: "Current instance inside class method.",
  super: "Access parent constructor/method.",
  if: "Conditional branch.",
  else: "Alternative branch.",
  while: "While loop.",
  for: "For loop.",
  break: "Break current loop.",
  continue: "Continue current loop.",
  try: "Start exception handling block.",
  catch: "Catch exception value.",
  finally: "Always executed block after try/catch.",
  throw: "Throw runtime exception.",
  return: "Return from function.",
  true: "Boolean true literal.",
  false: "Boolean false literal."
};

const builtins = [
  "print",
  "Promise.all",
  "Promise.race",
  "Promise.allSettled"
];

const nativeDocs = new Map(nativeSymbols.map((x) => [x.label, x.doc]));

connection.onInitialize((_params: InitializeParams): InitializeResult => {
  return {
    capabilities: {
      textDocumentSync: TextDocumentSyncKind.Incremental,
      completionProvider: {
        resolveProvider: false,
        triggerCharacters: ["."]
      },
      hoverProvider: true,
      definitionProvider: true
    }
  };
});

documents.onDidOpen((event) => {
  validateTextDocument(event.document);
});

documents.onDidChangeContent((event) => {
  validateTextDocument(event.document);
});

documents.onDidClose((event) => {
  connection.sendDiagnostics({ uri: event.document.uri, diagnostics: [] });
});

connection.onCompletion((_params): CompletionItem[] => {
  const keywordItems: CompletionItem[] = keywords.map((k) => ({
    label: k,
    kind: CompletionItemKind.Keyword
  }));

  const builtinItems: CompletionItem[] = builtins.map((b) => ({
    label: b,
    kind: CompletionItemKind.Function
  }));

  const nativeItems: CompletionItem[] = nativeSymbols.map((s) => ({
    label: s.label,
    kind: s.kind,
    detail: s.detail,
    documentation: s.doc
  }));

  return [...keywordItems, ...builtinItems, ...nativeItems];
});

connection.onHover((params: TextDocumentPositionParams): Hover | null => {
  const doc = documents.get(params.textDocument.uri);
  if (!doc) {
    return null;
  }

  const lineText = getLine(doc, params.position.line);
  const symbol = symbolAt(lineText, params.position.character);
  if (!symbol) {
    return null;
  }

  if (keywordDocs[symbol]) {
    return {
      contents: {
        kind: "markdown",
        value: `\`${symbol}\`  \n${keywordDocs[symbol]}`
      }
    };
  }

  if (nativeDocs.has(symbol)) {
    return {
      contents: {
        kind: "markdown",
        value: `\`${symbol}\`  \n${nativeDocs.get(symbol)}`
      }
    };
  }

  if (symbol === "Promise") {
    return {
      contents: {
        kind: "markdown",
        value: "`Promise` static helpers: `all`, `race`, `allSettled`."
      }
    };
  }

  return null;
});

connection.onDefinition((params: TextDocumentPositionParams): Location[] | null => {
  const doc = documents.get(params.textDocument.uri);
  if (!doc) {
    return null;
  }

  const lineText = getLine(doc, params.position.line);
  const word = wordAt(lineText, params.position.character);
  if (!word) {
    return null;
  }

  const defs = collectDefinitions(doc);
  const target = defs.get(word);
  if (!target) {
    return null;
  }

  return [
    Location.create(
      doc.uri,
      {
        start: target,
        end: { line: target.line, character: target.character + word.length }
      }
    )
  ];
});

async function validateTextDocument(textDocument: TextDocument): Promise<void> {
  const text = textDocument.getText();
  const diagnostics: Diagnostic[] = [];

  diagnostics.push(...validateDelimiters(textDocument, text));
  diagnostics.push(...validateSemicolons(textDocument, text));

  connection.sendDiagnostics({ uri: textDocument.uri, diagnostics });
}

function validateDelimiters(doc: TextDocument, text: string): Diagnostic[] {
  const diagnostics: Diagnostic[] = [];
  type StackItem = { ch: string; line: number; col: number };
  const stack: StackItem[] = [];

  let line = 0;
  let col = 0;
  let inString = false;
  let escaping = false;

  const pairs: Record<string, string> = {
    "{": "}",
    "[": "]",
    "(": ")"
  };

  const openSet = new Set(Object.keys(pairs));
  const closeSet = new Set(Object.values(pairs));

  for (let i = 0; i < text.length; i++) {
    const ch = text[i];
    const next = i + 1 < text.length ? text[i + 1] : "";

    if (ch === "\n") {
      line++;
      col = 0;
      escaping = false;
      continue;
    }

    if (!inString && ch === "/" && next === "/") {
      while (i < text.length && text[i] !== "\n") {
        i++;
      }
      i--;
      continue;
    }

    if (ch === "\"" && !escaping) {
      inString = !inString;
      col++;
      continue;
    }

    if (inString) {
      escaping = ch === "\\" && !escaping;
      col++;
      continue;
    }

    if (openSet.has(ch)) {
      stack.push({ ch, line, col });
    } else if (closeSet.has(ch)) {
      const top = stack.pop();
      if (!top) {
        diagnostics.push({
          severity: DiagnosticSeverity.Error,
          range: {
            start: { line, character: col },
            end: { line, character: col + 1 }
          },
          message: `Unmatched closing '${ch}'`,
          source: "ialang-lsp"
        });
      } else if (pairs[top.ch] !== ch) {
        diagnostics.push({
          severity: DiagnosticSeverity.Error,
          range: {
            start: { line, character: col },
            end: { line, character: col + 1 }
          },
          message: `Mismatched closing '${ch}', expected '${pairs[top.ch]}'`,
          source: "ialang-lsp"
        });
      }
    }

    escaping = false;
    col++;
  }

  while (stack.length > 0) {
    const top = stack.pop()!;
    const expected = pairs[top.ch];
    diagnostics.push({
      severity: DiagnosticSeverity.Error,
      range: {
        start: { line: top.line, character: top.col },
        end: { line: top.line, character: top.col + 1 }
      },
      message: `Unclosed '${top.ch}', expected '${expected}'`,
      source: "ialang-lsp"
    });
  }

  return diagnostics;
}

function validateSemicolons(doc: TextDocument, text: string): Diagnostic[] {
  const diagnostics: Diagnostic[] = [];
  const lines = text.split(/\r?\n/);
  const skipStart = /^(if|for|while|else|try|catch|finally|class|function|async function)\b/;
  const okEnd = /[;{}:,]$/;
  const likelyStatementEnd = /[\w\]\)\"']$/;

  for (let lineIndex = 0; lineIndex < lines.length; lineIndex++) {
    const raw = lines[lineIndex];
    const line = raw.trim();
    if (line.length === 0) {
      continue;
    }
    if (line.startsWith("//")) {
      continue;
    }
    if (skipStart.test(line)) {
      continue;
    }
    if (line.startsWith("import ") || line.startsWith("export ")) {
      continue;
    }
    if (okEnd.test(line)) {
      continue;
    }
    if (!likelyStatementEnd.test(line)) {
      continue;
    }

    const endChar = raw.length;
    diagnostics.push({
      severity: DiagnosticSeverity.Warning,
      range: {
        start: { line: lineIndex, character: Math.max(0, endChar - 1) },
        end: { line: lineIndex, character: endChar }
      },
      message: "Possible missing semicolon",
      source: "ialang-lsp"
    });
  }

  return diagnostics;
}

function getLine(doc: TextDocument, line: number): string {
  const all = doc.getText().split(/\r?\n/);
  if (line < 0 || line >= all.length) {
    return "";
  }
  return all[line];
}

function wordAt(lineText: string, character: number): string {
  if (character < 0 || character > lineText.length) {
    return "";
  }
  const left = lineText.slice(0, character);
  const right = lineText.slice(character);
  const leftMatch = left.match(/[A-Za-z_][A-Za-z0-9_]*$/);
  const rightMatch = right.match(/^[A-Za-z0-9_]*/);
  const l = leftMatch ? leftMatch[0] : "";
  const r = rightMatch ? rightMatch[0] : "";
  return l + r;
}

function symbolAt(lineText: string, character: number): string {
  if (character < 0 || character > lineText.length) {
    return "";
  }
  const left = lineText.slice(0, character);
  const right = lineText.slice(character);
  const leftMatch = left.match(/[A-Za-z_][A-Za-z0-9_.]*$/);
  const rightMatch = right.match(/^[A-Za-z0-9_.]*/);
  const l = leftMatch ? leftMatch[0] : "";
  const r = rightMatch ? rightMatch[0] : "";
  return l + r;
}

function collectDefinitions(doc: TextDocument): Map<string, Position> {
  const defs = new Map<string, Position>();
  const lines = doc.getText().split(/\r?\n/);

  // function foo(...) / async function foo(...) / export function foo(...)
  const functionRe = /^\s*(?:export\s+)?(?:async\s+)?function\s+([A-Za-z_][A-Za-z0-9_]*)\b/;
  // class Foo ... / export class Foo ...
  const classRe = /^\s*(?:export\s+)?class\s+([A-Za-z_][A-Za-z0-9_]*)\b/;
  // let x = ... / export let x = ...
  const letRe = /^\s*(?:export\s+)?let\s+([A-Za-z_][A-Za-z0-9_]*)\b/;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (line.trimStart().startsWith("//")) {
      continue;
    }

    for (const re of [functionRe, classRe, letRe]) {
      const m = line.match(re);
      if (!m) {
        continue;
      }
      const name = m[1];
      if (defs.has(name)) {
        continue;
      }
      const col = line.indexOf(name);
      if (col >= 0) {
        defs.set(name, Position.create(i, col));
      }
    }
  }

  return defs;
}

documents.listen(connection);
connection.listen();
