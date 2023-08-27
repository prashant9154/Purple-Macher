let socket = null;
let o = document.getElementById("output");
let userField = document.getElementById("username");
let personField = document.getElementById("person");
let messageField = document.getElementById("message");
let sendButton = document.getElementById("sendBtn");
let matchButton = document.getElementById("matchBtn");

let personMessageMap = new Map();

window.onbeforeunload = function () {
    console.log("Leaving");
    let jsonData = {};
    jsonData["action"] = "left";
    socket.send(JSON.stringify(jsonData))
}

document.addEventListener("DOMContentLoaded", function () {
    socket = new ReconnectingWebSocket("ws://127.0.0.1:8080/ws", null, { debug: true, reconectInterval: 3000 });

    const offline = `<span class="badge bg-danger">Not connected</span>`
    const online = `<span class="badge bg-success">Connected</span>`
    let statusDiv = document.getElementById("status");

    socket.onopen = () => {
        console.log("Successfully connected");
        statusDiv.innerHTML = online;
    }

    socket.onclose = () => {
        console.log("connection closed");
        statusDiv.innerHTML = offline;
    }

    socket.onerror = error => {
        console.log("there was an error");
        statusDiv.innerHTML = offline;
    }

    socket.onmessage = msg => {
        let data = JSON.parse(msg.data);
        console.log("Action is", data.action);

        switch (data.action) {
            case "list_users":
                let ul = document.getElementById("online_users");
                while (ul.firstChild) ul.removeChild(ul.firstChild);

                if (data.connected_users?.length > 0) {
                    data.connected_users.forEach(function (value) {
                        let li = document.createElement("li");
                        li.appendChild(document.createTextNode(value));
                        ul.appendChild(li);
                    })
                }
                break;

            case "broadcast":
                if (data.username != "" && data.person != "" && data.message != "") {
                    // Check if the key already exists in the map
                    if (personMessageMap.has(data.person)) {
                        // If the key exists, push the new message to the existing array
                        personMessageMap.get(data.person).push(data.message.toLowerCase());
                    } else {
                        // If the key doesn't exist, create a new array with the message
                        personMessageMap.set(data.person, [data.message.toLowerCase()]);
                    }
                    console.log(personMessageMap)
                }
                o.innerHTML = o.innerHTML + data.username + ":" + "guessed!" + "<br>";

                break;
        }
    }

    userField.addEventListener("change", function () {
        let jsonData = {};
        jsonData["action"] = "username";
        jsonData["username"] = this.value;
        socket.send(JSON.stringify(jsonData));
    })

    messageField.addEventListener("keydown", function (event) {
        if (event.code === "Enter") {
            if (!socket) {
                console.log("no connection");
                return false
            }

            if ((userField.value === "") || (messageField.value === "")) {
                errorMessage("Fill out username and message!");
                return false;
            } else {
                sendMessage()
            }

            event.preventDefault();
            event.stopPropagation();
        }
    })

    sendButton.addEventListener("click", function () {
        if ((userField.value === "") || (messageField.value === "")) {
            errorMessage("Fill out username and message!");
            return false;
        } else {
            sendMessage()
        }
    })

    matchButton.addEventListener("click", function () {
        if (personMessageMap.size >= 1) {
            const valuesArray = personMessageMap.get(personField.value)
            const lastTwoValues = valuesArray.slice(-2);
            let compare = ""

            for (const value of lastTwoValues) {
                if (compare === "") {
                    console.log(value);
                    compare = value;
                } else if (compare !== value) {
                    console.log(value);
                    o.innerHTML = o.innerHTML + "<span style='color: red;'>Not Matched! <br></span>";
                    return
                }
            }

            o.innerHTML = o.innerHTML + "<span style='color: green;'>Matched! <br></span>";
        }
        console.log(personMessageMap.size)
    })
})

function sendMessage() {
    let jsonData = {};
    jsonData["action"] = "broadcast";
    jsonData["username"] = userField.value;
    jsonData["person"] = personField.value;
    jsonData["message"] = messageField.value;
    socket.send(JSON.stringify(jsonData))
    messageField.value = "";
}

function errorMessage(msg) {
    notie.alert({
        type: 'error',
        text: msg,
    })
}