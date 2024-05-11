import './App.css'
import { ContactGroup } from "./ContactGroup"
import { ChatGroup } from "./ChatGroup"
import { createSignal, onMount } from 'solid-js';
import { ChatRoom, TurboGuacClient } from "turbosdk-js"


function App() {
  let tgc: TurboGuacClient | null;
  let [onlineUsers, setOnlineUsers] = createSignal<string[]>();
  let [activeChatRooms, setActiveChatRooms] = createSignal<ChatRoom[]>();

  async function tryConnect(serverURL: string, userName: string) {
    if (tgc) return true;
    try {
      tgc = await TurboGuacClient.createClient(serverURL, userName);
      let onlineUzers = await tgc.getOnlineUsers();
      setOnlineUsers(onlineUzers);
      let activeRooms = await tgc.getMyChatRooms();
      setActiveChatRooms(activeRooms);
    }
    catch (e) {
      console.log("Erro while establishing websocket connection", e);
      tgc = null;
      return false;
    }
    return true;
  }

  onMount(async function() {
    const modal = document.getElementById('input_modal') as HTMLDialogElement;
    modal.showModal();
    modal.addEventListener("close", async function(_e) {
      const serverUrl = (document.getElementById('server-addr') as HTMLInputElement).value;
      const userName = (document.getElementById('username') as HTMLInputElement).value;
      if (await tryConnect(serverUrl, userName)) {
      } else {
        modal.showModal();
      }
    });
  });

  return (
    <>
      <div class="navbar app-header">
        <p class="text-2xl">🐌 Turbo Guac 🥑</p>
      </div>
      <dialog id="input_modal" class="modal">
        <div class="modal-box">
          <h3 class="font-bold text-lg">Enter Chat Server address</h3>
          <p class="py-4">Enter Chat Server address</p>
          <label class="input input-bordered flex items-center gap-2">
            <input id="server-addr" type="text" class="grow" placeholder="https://localhost:8080" />
          </label>
          <p class="py-4">Enter username</p>
          <label class="input input-bordered flex items-center gap-2">
            <input id="username" type="text" class="grow" placeholder="" />
          </label>
          <div class="modal-action">
            <form method="dialog">
              <button class="btn">Enter</button>
            </form>
          </div>
        </div>
      </dialog>
      <div class="app-body">
        <ContactGroup onlineUsers={onlineUsers()} myChatRooms={activeChatRooms()} />
        <ChatGroup />
      </div>
    </>
  )
}

export default App
