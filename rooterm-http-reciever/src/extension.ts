import * as vscode from "vscode";
import * as http from "http";
import { RooCodeAPI } from "@roo-code/types";

const getRooCodeAPI = (): RooCodeAPI => {
  const rooCodeExtension = vscode.extensions.getExtension<RooCodeAPI>(
    "RooVeterinaryInc.roo-cline",
  );

  if (!rooCodeExtension?.isActive) {
    throw new Error("Extension is not activated");
  }

  const rooCodeApi: RooCodeAPI = rooCodeExtension.exports;

  if (!rooCodeApi) {
    throw new Error("API is not available");
  }

  return rooCodeApi;
};

function initServer(context: vscode.ExtensionContext) {
  const config = vscode.workspace.getConfiguration("httpReceiver");
  const port: number = config.get("port", 9421);

  const outputChannel = vscode.window.createOutputChannel(
    "RooTerm HTTP Receiver",
  );

  outputChannel.appendLine(
    `RooTerm HTTP Receiver is starting on port ${port}...`,
  );

  const rooCodeApi = getRooCodeAPI();

  // FIXME:
  rooCodeApi.on("message", (event: unknown) => {
    outputChannel.appendLine("Received message from RooCode:");
    outputChannel.appendLine(JSON.stringify(event, null, 2));
  });

  // Create RooTerm HTTP server
  const server = http.createServer((req, res) => {
    if (req.method === "POST" && req.url === "/") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        try {
          const data = JSON.parse(body) as { message: string; time: string };
          // Show the received message in VSCode
          vscode.window.showInformationMessage(
            `Received at ${data.time}: ${data.message}`,
          );
          res.writeHead(200, { "Content-Type": "text/plain" });
          res.end("OK");

          rooCodeApi.sendMessage(data.message);
        } catch (err) {
          console.error("Failed to parse JSON", err);
          res.writeHead(400, { "Content-Type": "text/plain" });
          res.end("Invalid JSON");
        }
      });
    } else {
      res.writeHead(404);
      res.end();
    }
  });

  server.on("error", (err) => {
    vscode.window.showErrorMessage(`RooTerm HTTP Server error: ${err.message}`);
  });

  server.listen(port, "127.0.0.1", () => {
    console.log(`RooTerm HTTP Receiver listening on http://127.0.0.1:${port}/`);
  });

  // Dispose server on extension deactivation
  context.subscriptions.push({
    dispose: () => {
      server.close();
    },
  });
}

export function activate(context: vscode.ExtensionContext) {
  vscode.commands.registerCommand("rooterm-http-reciever.start-server", () =>
    initServer(context),
  );
}

export function deactivate() {}
