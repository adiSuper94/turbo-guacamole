import { For, mergeProps } from "solid-js";
import { ChatRoom } from "turbosdk-js";

interface Props {
  onlineUsers: string[] | undefined;
  myChatRooms: () => ChatRoom[] | undefined;
  activeChatRoom: () => ChatRoom | undefined;
  setActiveChatRoom: (s: ChatRoom | undefined) => void
}


export function ContactGroup(propArgs: Props) {
  const props = mergeProps({ onlineUsers: [], myChatRooms: [] }, propArgs);
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
                <tr id={`online-users-${i}`} class="hover">
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
