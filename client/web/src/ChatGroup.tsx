import { For, Match, Switch, createEffect, createSignal } from "solid-js"
import { Message, TurboGuacClient } from "turbosdk-js";

interface Props {
  tgc: () => TurboGuacClient | undefined;
  chatRoomId: () => string | undefined;
}

const [messages, setMessages] = createSignal<Message[]>();

export function ChatGroup(props: Props) {
  const tgc = props.tgc;
  createEffect(async function() {
    let roomId = props.chatRoomId();
    if (roomId != null && props.tgc() != null) {
      const messagez = await tgc()!.getMessages(roomId!);
      setMessages(messagez);
    }
  });


  return (
    <>
      <div class="chat-group">
        <div class="chats overflow-x-auto">
          <For each={messages()}>{(message, _i) =>
            <Switch fallback={
              <div class="chat chat-start">
                <div class="chat-bubble">{message.Body}</div>
              </div>
            }>
              <Match when={message.SenderID == tgc()?.getUserName()??""}>
                <div class="chat chat-end">
                  <div class="chat-bubble">{message.Body}</div>
                </div>
              </Match>
            </Switch>
          }
          </For >
        </div >
        <div class="message-input join">
          <textarea class="textarea textarea-bordered join-item"></textarea>
          <button class="btn join-item">Send</button>
        </div>
      </div >
    </>
  );
}
