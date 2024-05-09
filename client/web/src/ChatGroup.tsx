function ChatGroup(){
  return (
    <>
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
    </>
  );
}
export default ChatGroup;
