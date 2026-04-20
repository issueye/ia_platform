import * as path from "path";
import * as vscode from "vscode";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind
} from "vscode-languageclient/node";

let client: LanguageClient | undefined;

export function activate(context: vscode.ExtensionContext): void {
  const serverModule = context.asAbsolutePath(path.join("out", "server.js"));

  const serverOptions: ServerOptions = {
    run: {
      module: serverModule,
      transport: TransportKind.ipc
    },
    debug: {
      module: serverModule,
      transport: TransportKind.ipc,
      options: {
        execArgv: ["--nolazy", "--inspect=6009"]
      }
    }
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: "file", language: "ialang" }],
    synchronize: {
      fileEvents: vscode.workspace.createFileSystemWatcher("**/*.ia")
    }
  };

  client = new LanguageClient(
    "ialangLanguageServer",
    "ialang Language Server",
    serverOptions,
    clientOptions
  );

  context.subscriptions.push(client);
  void client.start();
}

export async function deactivate(): Promise<void> {
  if (!client) {
    return;
  }
  await client.stop();
}
