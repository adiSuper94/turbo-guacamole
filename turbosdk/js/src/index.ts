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
  Username: string;
  ChatRoomID: string;
}

export type Message = {
  ID: string;
  Body: string;
  ChatRoomID: string;
  SenderID: string;
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

  onMessage(fn: (wsMsg: WSMessage) => void) {
    this.ws.onmessage = ((event: MessageEvent) => {
      const wsMsg = JSON.parse(event.data);
      fn(wsMsg);
    });
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

  async createChatRoom(roomName: string): Promise<ChatRoom> {
    const url = `http://${this.serverAddr}/chatrooms?username=${this.userName}&chatroom_name=${roomName}`;
    let response = await fetch(url, { method: "POST", headers: { "Accept": "application/json" } });
    return await response.json();
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

  async startDM(username: string): Promise<DM|null> {
    if(username.trim() == this.userName.trim()) return null;
    const dms: DM[] = await this.getDMs() ?? [];
    for (const dm of dms) {
      if (dm.Username == username) {
        return dm;
      }
    }
    const chatRoom = await this.createChatRoom(encodeURIComponent(`${username}&${this.userName}`));
    this.addMemberToChatRoom(chatRoom.ID, username);
    return { Username: username, ChatRoomID: chatRoom.ID };
  }

  addMemberToChatRoom(chatRoomId: string, userName: string) {
    const addMemberRequest: WSMessage = {
      Id: crypto.randomUUID(),
      Type: WSMessageType.AddMemberToChatRoom,
      Data: userName,
      To: chatRoomId,
      From: this.userName
    };
    this.sendWSMessage(addMemberRequest);
  }

  async getDMs() {
    const url = `http://${this.serverAddr}/dms?username=${this.userName}`;
    let response = await fetch(url, { method: "GET", headers: { "Accept": "application/json" } });
    let data: DM[] = await response.json();
    return data;
  }

  async getMessages(chatRoomId: string) {
    if (chatRoomId.trim() == "" || chatRoomId == NilUUID) return [];
    const url = `http://${this.serverAddr}/messages?chatRoomId=${chatRoomId}`;
    let response = await fetch(url, { method: "GET", headers: { "Accept": "application/json" } });
    let data: Message[] = await response.json();
    return data;
  }
}
