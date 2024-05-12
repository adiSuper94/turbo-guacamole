"use strict";

export enum WSMessageType {
  Text = 0,
  Login,
  LoginAck,
  Logout,
  SingleTick,
  AddMemberToChatRoom
}

export type WSMessage = {
  Id: string;
  Data: string;
  To: string;
  From: string;
  Type: WSMessageType;
}

export type ChatRoom = {
  ID: string;
  Name: string;
  CreatedAt: Date
  ModifiedAt: Date
}

export type DM = {
  UserName: string;
  ChatRoomId: string;
}

export type Message = {
  Id: string;
  Body: string;
  ChatRoomId: string;
  SenderID: string;
  CreatedAt: Date;
  ModifiedAt: Date;
}

export const NilUUID = "00000000-0000-0000-0000-000000000000";

export class TurboGuacClient {
  private ws: WebSocket
  private userName: string
  private serverAddr: string

  private constructor(serverAddr: string, userName: string) {
    this.serverAddr = serverAddr;
    this.userName = userName;
    this.ws = new WebSocket("ws://" + serverAddr + "/ws");
  }

  static async createClient(serverAddr: string, userName: string) {
    const tgc = new TurboGuacClient(serverAddr, userName);
    let p = new Promise((resolve, reject) => {
      tgc.ws.addEventListener("open", resolve, { once: true });
      tgc.ws.addEventListener("error", reject, { once: true });
    });
    await p;
    tgc.loginOrRegister();
    return tgc;
  }

  getUserName() { return this.userName; }

  private loginOrRegister() {
    const loginRequest: WSMessage = {
      Id: crypto.randomUUID(),
      Data: "login",
      To: NilUUID,
      Type: WSMessageType.Login,
      From: this.userName
    };
    this.sendWSMessage(loginRequest);
  }

  private sendWSMessage(message: WSMessage) {
    this.ws.send(JSON.stringify(message));
  }

  async sendMessage(data: string, chatRoomId: string) {
    if (chatRoomId.trim() == "") {
      return null;
    }
    const message: WSMessage = {
      Id: crypto.randomUUID(),
      Data: data,
      To: chatRoomId,
      From: this.userName,
      Type: WSMessageType.Text
    }
    this.sendWSMessage(message);
  }

  async createChatRoom(roomName: string) {
    const url = `http://${this.serverAddr}/chatrooms?username=${this.userName}&chatroom_name=${roomName}`;
    let response = await fetch(url, { method: "POST", headers: { "Accept": "application/json" } });
    return response;
  }

  async getMyChatRooms() {
    const url = `http://${this.serverAddr}/chatrooms?username=${this.userName}`;
    let response = await fetch(url, { method: "GET", headers: { "Accept": "application/json" } });
    let data: ChatRoom[] = await response.json();
    return data;

  }

  async getOnlineUsers(): Promise<string[]> {
    const url = `http://${this.serverAddr}/online-users`;
    let response = await fetch(url, { method: "GET", headers: { "Accept": "application/json" } });
    const data = await response.json();
    return data;
  }

  async getDMs() {
    const url = `http://${this.serverAddr}/dms?username=${this.userName}`;
    let response = await fetch(url, { method: "GET", headers: { "Accept": "application/json" } });
    let data: DM[] = await response.json();
    return data;
  }

  async getMessages(chatRoomId: string) {
    const url = `http://${this.serverAddr}/messages?chatRoomId=${chatRoomId}`;
    let response = await fetch(url, { method: "GET", headers: { "Accept": "application/json" } });
    let data: Message[] = await response.json();
    return data;
  }
}
