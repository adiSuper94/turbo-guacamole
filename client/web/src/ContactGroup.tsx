function ContactGroup() {
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
              <tr class="hover">
                <td>Prajit</td>
              </tr>
              <tr class="hover">
                <td>Arun</td>
              </tr>
              <tr class="hover">
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
              <tr class="hover bg-base-200">
                <td>Batman</td>
              </tr>
              <tr class="hover">
                <td>SV & Hari & Adi</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

export default ContactGroup
