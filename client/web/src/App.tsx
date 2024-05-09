import './App.css'
import ContactGroup from "./ContactGroup"
import { ChatGroup } from "./ChatGroup"
import { onMount } from 'solid-js';
import { TurboGuacClient } from "turbosdk-js"


function App() {
  let tgc: TurboGuacClient;
  let serverUrl;

  function tryConnect(serverURL: string, userName: string): boolean {
    if (tgc) return true;
    serverUrl = serverURL;
    tgc = new TurboGuacClient(serverUrl, userName);
    return true;
  }

  onMount(function() {
    const modal = document.getElementById('input_modal') as HTMLDialogElement;
    modal.showModal();
    console.log('modal', modal);
    modal.addEventListener("close", function(e) {
      console.log('closed', e);
      const serverUrl = (document.getElementById('server-addr') as HTMLInputElement).value;
      const userName = (document.getElementById('username') as HTMLInputElement).value;
      if (tryConnect(serverUrl, userName)) {
        console.log('connected');
      } else {
        console.log('not connected');
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
        <ContactGroup />
        <ChatGroup />
      </div>
    </>
  )
}

export default App
