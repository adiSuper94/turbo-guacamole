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
  id: string;
  data: string;
  to: string;
  from: string;
  type: WSMessageType;
}

export type ChatRoom = {
  id: string;
  name: string;
  createdAt: Date
  modifiedAr: Date
}

export type DM = {
  userName: string;
  chatRoomId: string;
}

export class TurboGuacClient {
  private ws: WebSocket
  private userName: string
  private serverAddr: string

  constructor(serverAddr: string, userName: string) {
    this.serverAddr = serverAddr;
    this.userName = userName;
    this.ws = new WebSocket("ws://" + serverAddr + "/ws");
  }
  async createChatRoom(roomName: string) {
    const url = `http://${this.serverAddr}/chatrooms?username=${this.userName}&chatroom_name=${roomName}`;
    let response = await fetch(url, { method: "POST", headers: { "Accept": "application/json" } });
    return response;
  }

  async getMyChatRooms() {
    const url = `http://${this.serverAddr}/chatrooms?username=${this.userName}`;
    let response = await fetch(url, { method: "GET", headers: { "Accept": "application/json" } });
    return response;
  }

  async getOnlineUsers() {
    const url = `http://${this.serverAddr}/online-users`;
    let response = await fetch(url, { method: "GET", headers: { "Accept": "application/json" } });
    return response;
  }

  async getDMs() {
    const url = `http://${this.serverAddr}/dms?username=${this.userName}`;
    let response = await fetch(url, { method: "GET", headers: { "Accept": "application/json" } });
    return response;
  }
}
