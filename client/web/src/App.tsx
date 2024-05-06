import './App.css'

function App() {

  return (
    <>
    <div class="navbar app-header">
      <p class="text-2xl">🐌 Turbo Guac 🥑</p>
    </div>
    <div class="app-body">
      <div class="contact-group">
        <div class="online-users overflow-x-auto">
          <table class="table">
            <thead>
              <tr>
                <th style="background:rgba(0, 0, 0, 0.3)">Online Users</th>
              </tr>
            </thead>
            <tbody>
              <tr class = "hover">
                <td>Prajit</td>
              </tr>
              <tr class="hover">
                <td>Arun</td>
              </tr>
              <tr class = "hover">
                <td>Haroon</td>
              </tr>
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
              <tr class = "hover bg-base-200">
                <td>Batman</td>
              </tr>
              <tr class="hover">
                <td>SV & Hari & Adi</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div class="chat-group">
        <div class="chats overflow-x-auto">
          <div class="chat chat-start">
            <div class="chat-bubble">It's over Anakin, <br/>I have the high ground.</div>
          </div>
          <div class="chat chat-end">
            <div class="chat-bubble">You underestimate my power!</div>
          </div>
        </div>
        <div class="message-input join">
          <textarea class = "textarea textarea-bordered join-item"></textarea>
          <button class = "btn join-item">Send</button>
        </div>
      </div>
    </div>
    </>
  )
}

export default App
