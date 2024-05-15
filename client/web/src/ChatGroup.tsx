import { For, Match, Switch, createEffect, createSignal } from "solid-js"
import { Message, NilUUID, TurboGuacClient, WSMessage, WSMessageType } from "turbosdk-js";

interface Props {
  tgc: () => TurboGuacClient | undefined;
  chatRoomId: () => string | undefined;
}

const [messages, setMessages] = createSignal<Message[]>([], { equals: false });

async function sendMessage(props: Props) {
  const textArea = (document.getElementById("textarea")) as HTMLTextAreaElement;
  const text = textArea.value;
  textArea.value = "";
  if (text == null || text.trim() == "") return;
  if (props.tgc() == null) return;
  const chatRoomId = props.chatRoomId();
  if (chatRoomId == null) return;
  const userName = props.tgc()!.getUserName();
  let message: Message = {
    ID: NilUUID,
    Body: text,
    SenderID: userName,
    ChatRoomID: chatRoomId
  };
  let currMessages = messages() ?? [];
  currMessages.push(message);
  setMessages(currMessages);
  await props.tgc()!.sendMessage(text, chatRoomId);
  const messageBox = document.getElementById("message-list")!
  messageBox.scrollTop = messageBox.scrollHeight;
}

export function ChatGroup(props: Props) {
  const tgc = props.tgc;
  createEffect(async function() {
    let roomId = props.chatRoomId();
    if (roomId != null && props.tgc() != null) {
      const messagez = await tgc()!.getMessages(roomId!);
      tgc()?.onMessage(processIncomingMsg);
      setMessages(messagez);
    }
  });

  function processIncomingMsg(wsMsg: WSMessage) {
    console.log(JSON.stringify(wsMsg, null, 2));
    if (wsMsg.Type != WSMessageType.Text) return;
    if (wsMsg.To != props.chatRoomId()) return;
    const msg: Message = {
      ID: wsMsg.Id,
      Body: wsMsg.Data,
      SenderID: wsMsg.From,
      ChatRoomID: wsMsg.To
    };
    let currMessages = messages() ?? [];
    currMessages?.push(msg);
    setMessages(currMessages);
    const messageBox = document.getElementById("message-list")!
    messageBox.scrollTop = messageBox.scrollHeight;
  }


  return (
    <>
      <div class="chat-group">
        <div id="message-list" class="chats overflow-x-auto">
          <For each={messages()}>{(message, _i) =>
            <Switch fallback={
              <div class="chat chat-start">
                <div class="chat-bubble">{message.Body}</div>
              </div>
            }>
              <Match when={message.SenderID == tgc()?.getUserName()}>
                <div class="chat chat-end">
                  <div class="chat-bubble">{message.Body}</div>
                </div>
              </Match>
            </Switch>
          }
          </For >
        </div >
        <div class="message-input join">
          <textarea id="textarea" class="textarea textarea-bordered join-item" onKeyUp={async (event) => { if (event.keyCode == 13) await sendMessage(props) }}></textarea>
          <button class="btn join-item" onClick={async () => await sendMessage(props)}>Send</button>
        </div>
      </div >
    </>
  );
}
