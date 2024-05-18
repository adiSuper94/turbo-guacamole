#  🚧 WIP
# turbo-guacamole
## ௳
An app that lets you chat with people near you, and doesn't let you if you aren't near them.

## Development setup
- Create Postgres DB named `turbo`
- In `./server/db.go` update the func createDBConnection with the correct DB URL.
- Run `./server/turbo.sql` on `turbo` db to initialize the db schema.
- cd into `./server` and run `go run .` to start the server.
- To start the web client :
    - cd into `./turbosdk/js`
    - run `npm install` to install dependencies
    - run `npm run build` to build the turbo-js sdk
    - cd into `./client/web/`
    - run `npm install` to install dependencies
    - run `npm run dev` to start the client app
- To start the cli client:
    - cd into `./client/go-cli`
    - run `go run .`
