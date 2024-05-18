import { For, mergeProps } from "solid-js";
import { ChatRoom, TurboGuacClient } from "turbosdk-js";

interface Props {
  onlineUsers: string[] | undefined;
  myChatRooms: () => ChatRoom[] | undefined;
  setMyChatRooms: (rooms: ChatRoom[] | undefined) => void;
  activeChatRoom: () => ChatRoom | undefined;
  setActiveChatRoom: (s: ChatRoom | undefined) => void
  tgc: () => TurboGuacClient | undefined;
}


export function ContactGroup(propArgs: Props) {
  const props: Props = mergeProps({ onlineUsers: []}, propArgs);
  function onClickMyChat(newId: string) {
    let oldActive = document.getElementById(props.activeChatRoom()?.ID ?? "");
    if (oldActive) {
      oldActive.className = "hover";
    }
    let newActive = document.getElementById(newId)!;
    newActive.className = "hover bg-base-200";
    const newChatRoom = props.myChatRooms()?.find(chatRoom => chatRoom.ID == newId);
    props.setActiveChatRoom(newChatRoom);

  }
  async function onClickOnlineUsers(onlineUser: string) {
    const tgc = props.tgc()!;
    const dm = await tgc.startDM(onlineUser);
    console.log(dm);
    if (dm == null) return;
    const existingChatRoom = props.myChatRooms()?.find(chatRoom => chatRoom.ID == dm?.ChatRoomID);
    if (existingChatRoom != null) {
      console.log(existingChatRoom);
      props.setActiveChatRoom(existingChatRoom);
      return;
    }
    const myChatRooms = await tgc.getMyChatRooms();
    props.setMyChatRooms(myChatRooms);

  }
  return (
    <>
      <div class="contact-group">
        <div class="online-users overflow-x-auto">
          <table class="table">
            <thead>
              <tr>
                <th style="background:rgba(0, 0, 0, 0.3)">Online Users</th>
              </tr>
            </thead>
            <tbody>
              <For each={props.onlineUsers}>{(onlineUser, i) =>
                <tr id={`online-users-${i}`} class="hover" onClick={async () => await onClickOnlineUsers(onlineUser)}>
                  <td>{onlineUser}</td>
                </tr>
              }
              </For>
            </tbody>
          </table>
        </div>
        <div class="chat-rooms overflow-x-auto">
          <table class="table">
            <thead>
              <tr>
                <th style="background:rgba(0, 0, 0, 0.3)">Chat Rooms</th>
              </tr>
            </thead>
            <tbody>
              <For each={props.myChatRooms()}>{(myChatRoom, _i) =>
                <tr id={myChatRoom.ID} class="hover" onClick={() => onClickMyChat(myChatRoom.ID)}>
                  <td>{myChatRoom.Name}</td>
                </tr>
              }
              </For>
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

export default ContactGroup
