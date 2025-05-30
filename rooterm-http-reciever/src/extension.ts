import * as vscode from "vscode";
import * as http from "http";
import { ClineMessage, RooCodeAPI } from "@roo-code/types";

interface Payload {
  message: string;
  time: string;
}

function getRooCodeAPI(): RooCodeAPI {
  const extension = vscode.extensions.getExtension<RooCodeAPI>(
    "RooVeterinaryInc.roo-cline",
  );
  if (!extension) {
    throw new Error("RooCode extension is not installed");
  }
  if (!extension.isActive) {
    throw new Error("RooCode extension is not activated");
  }
  const api = extension.exports;
  if (!api) {
    throw new Error("RooCode API is not available");
  }
  return api;
}

function subscribeToRooCodeMessages(
  api: RooCodeAPI,
  channel: vscode.OutputChannel,
): void {
  api.on(
    "message",
    (event: {
      taskId: string;
      action: "created" | "updated";
      message: ClineMessage;
    }) => {
      channel.appendLine(
        `(ID: ${event.taskId}) Received message from RooCode:`,
      );
      channel.appendLine(JSON.stringify(event.message, null, 2));
    },
  );
}

function readRequestBody(req: http.IncomingMessage): Promise<string> {
  return new Promise((resolve, reject) => {
    let body = "";
    req.on("data", (chunk) => {
      body += chunk;
    });
    req.on("end", () => {
      resolve(body);
    });
    req.on("error", (err) => {
      reject(err);
    });
  });
}

function initServer(context: vscode.ExtensionContext): void {
  const config = vscode.workspace.getConfiguration("httpReceiver");
  const port = config.get<number>("port", 9421);
  const channel = vscode.window.createOutputChannel("RooTerm HTTP Receiver");
  channel.appendLine(`RooTerm HTTP Receiver is starting on port ${port}...`);

  let api: RooCodeAPI;
  try {
    api = getRooCodeAPI();
    subscribeToRooCodeMessages(api, channel);
  } catch (err) {
    const message =
      err instanceof Error
        ? err.message
        : "Unknown error during API initialization";
    vscode.window.showErrorMessage(
      `Failed to initialize RooCode API: ${message}`,
    );
    return;
  }

  const server = http.createServer(async (req, res) => {
    if (req.method === "POST" && req.url === "/") {
      try {
        const body = await readRequestBody(req);
        const data = JSON.parse(body) as Payload;
        channel.appendLine(`Received at ${data.time}: ${data.message}`);
        vscode.window.showInformationMessage(
          `Received at ${data.time}: ${data.message}`,
        );
        api.sendMessage(data.message);
        res.writeHead(200, { "Content-Type": "text/plain" });
        res.end("OK");
      } catch (err) {
        const message = err instanceof Error ? err.message : "Unknown error";
        channel.appendLine(`Error handling request: ${message}`);
        console.error(err);
        res.writeHead(400, { "Content-Type": "text/plain" });
        res.end("Invalid JSON");
      }
    } else {
      res.writeHead(404);
      res.end();
    }
  });

  server.on("error", (err) => {
    vscode.window.showErrorMessage(`RooTerm HTTP Server error: ${err.message}`);
  });

  server.listen(port, "127.0.0.1", () => {
    channel.appendLine(
      `RooTerm HTTP Receiver listening on http://127.0.0.1:${port}/`,
    );
  });

  context.subscriptions.push({ dispose: () => server.close() });
}

export function activate(context: vscode.ExtensionContext): void {
  const disposable = vscode.commands.registerCommand(
    "rooterm-http-reciever.start-server",
    () => {
      initServer(context);
    },
  );
  context.subscriptions.push(disposable);
}

export function deactivate(): void {}
