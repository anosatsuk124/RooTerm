import { WebSocketServer } from 'ws';
import { OutputChannel, ExtensionContext } from 'vscode';
import { RooCodeAPI } from '@roo-code/types';
import { RequestHandler } from './RequestHandler';

export class HttpServer {
  private wss: WebSocketServer;

  constructor(
    private port: number,
    private host: string,
    private api: RooCodeAPI,
    private channel: OutputChannel,
    private context: ExtensionContext
  ) {
    this.wss = new WebSocketServer({ port, host });
    this.setupServerHandlers();
  }

  private setupServerHandlers(): void {
    this.wss.on('connection', (socket) => {
      new RequestHandler(this.wss, this.api, this.channel);
    });

    this.wss.on('error', (err) => {
      this.channel.appendLine(`WebSocket Server error: ${err.message}`);
    });

    this.wss.on('listening', () => {
      this.channel.appendLine(
        `RooTerm WebSocket Receiver listening on ws://${this.host}:${this.port}/`
      );
    });

    this.context.subscriptions.push({ dispose: () => this.wss.close() });
  }

  public start(): WebSocketServer {
    return this.wss;
  }
}