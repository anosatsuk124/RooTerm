import { WebSocketServer } from 'ws';
import { OutputChannel } from 'vscode';
import { RooCodeAPI } from '@roo-code/types';

interface Payload {
  message: string;
  time: string;
}

export class RequestHandler {
  constructor(
    private wss: WebSocketServer,
    private api: RooCodeAPI,
    private channel: OutputChannel
  ) {
    this.setupMessageHandlers();
  }

  private setupMessageHandlers(): void {
    this.wss.on('connection', (socket) => {
      socket.on('message', (data) => {
        try {
          const { message, time } = JSON.parse(data.toString()) as Payload;
          this.channel.appendLine(`Received at ${time}: ${message}`);
          this.api.sendMessage(message);
        } catch (err) {
          this.channel.appendLine(
            `Error handling message: ${err instanceof Error ? err.message : err}`
          );
          socket.send('Invalid JSON');
        }
      });
    });
  }
}