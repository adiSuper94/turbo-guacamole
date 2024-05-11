import { For, mergeProps } from "solid-js";
import { ChatRoom } from "turbosdk-js";

interface Props {
  onlineUsers: string[] | undefined;
  myChatRooms: ChatRoom[] | undefined;
}
export function ContactGroup(propArgs: Props) {
  const props = mergeProps({ onlineUsers: [], myChatRooms: [] }, propArgs);
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
              <For each={props.onlineUsers}>{(onlineUser, _i) =>
                <tr class="hover">
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
              <For each={props.myChatRooms}>{(myChatRoom, _i) =>
                <tr class="hover bg-base-200">
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
