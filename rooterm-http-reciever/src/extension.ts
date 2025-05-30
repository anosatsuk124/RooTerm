import * as vscode from "vscode";
import { WebSocketServer } from "ws";
import { RooCodeAPI } from "@roo-code/types";

interface Payload {
  message: string;
  time: string;
}

interface ChatPayload {
  is_reasoning: boolean;
  message: string;
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
  wss: WebSocketServer,
): void {
  api.on("message", (event) => {
    channel.appendLine(`(ID: ${event.taskId}) Received message from RooCode:`);
    channel.appendLine(JSON.stringify(event.message, null, 2));
    wss.clients.forEach((client) => {
      if (client.readyState === client.OPEN) {
        if (event.message.text) {
          const isReasoning = event.message.say === "reasoning";
          const payload: ChatPayload = {
            is_reasoning: isReasoning,
            message: event.message.text,
          };
          client.send(JSON.stringify(payload));
        }
      }
    });
  });
}

function setupWebSocketServer(
  api: RooCodeAPI,
  channel: vscode.OutputChannel,
  context: vscode.ExtensionContext,
  port: number,
): void {
  const wss = new WebSocketServer({ port, host: "127.0.0.1" });
  subscribeToRooCodeMessages(api, channel, wss);

  wss.on("connection", (socket) => {
    socket.on("message", (data) => {
      try {
        const { message, time } = JSON.parse(data.toString()) as Payload;
        channel.appendLine(`Received at ${time}: ${message}`);
        api.sendMessage(message);
      } catch (err) {
        channel.appendLine(
          `Error handling message: ${err instanceof Error ? err.message : err}`,
        );
        socket.send("Invalid JSON");
      }
    });
  });

  wss.on("error", (err) => {
    vscode.window.showErrorMessage(`WebSocket Server error: ${err.message}`);
  });

  wss.on("listening", () => {
    channel.appendLine(
      `RooTerm WebSocket Receiver listening on ws://127.0.0.1:${port}/`,
    );
  });

  context.subscriptions.push({ dispose: () => wss.close() });
}

function initServer(context: vscode.ExtensionContext): void {
  const config = vscode.workspace.getConfiguration("httpReceiver");
  const port = config.get<number>("port", 9421);
  const channel = vscode.window.createOutputChannel(
    "RooTerm WebSocket Receiver",
  );

  let api: RooCodeAPI;
  try {
    api = getRooCodeAPI();
    setupWebSocketServer(api, channel, context, port);
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
