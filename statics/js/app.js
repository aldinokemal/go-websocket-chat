const DB = idb.openDB("chatroom", 2, {
    upgrade(db) {
        db.createObjectStore("messages")
        db.createObjectStore("account")
    },
})

const dbMessage = {
    async get(key) {
        return (await DB).get('messages', key);
    },
    async set(key, val) {
        return (await DB).put('messages', val, key);
    },
    async delete(key) {
        return (await DB).delete('messages', key);
    },
    async clear() {
        return (await DB).clear('messages');
    },
    async keys() {
        return (await DB).getAllKeys('messages');
    },
};

const dbAccount = {
    async get(key) {
        return (await DB).get('account', key);
    },
    async set(key, val) {
        return (await DB).put('account', val, key);
    },
    async delete(key) {
        return (await DB).delete('account', key);
    },
    async clear() {
        return (await DB).clear('account');
    },
    async keys() {
        return (await DB).getAllKeys('account');
    },
};

new Vue({
    el: "#vueapp",
    delimiters: ['{%', '%}'],
    data: {
        ws_channel: "chatroom",
        ws_data: {
            channel: "chatroom",
            event: "",
            message: {}
        },
        toastMessage: '',
        myUUID: '',
        myName: '',
        isWebsocketExist: false,
        dataMessage: [],
        textMessage: '',
        onlineTotal: 0,
        onlineData: [],
        conn: null,
        previewImageName: '',
        previewImageURL: '',
        idb: null
    },
    created() {
        dbAccount.get(1).then(async data => {
            if (data != null) {
                this.myName = data.name
                this.myUUID = data.uuid
                setTimeout(() => {
                    this.settingUpWebSocket()
                    this.settingUpMessages()
                }, 200)
            }
        })
        setTimeout(() => $("#username").focus(), 500)
    },
    methods: {
        settingUpMessages() {
            DB.then(async db => {
                let cursor = await db.transaction("messages").store.openCursor();

                while (cursor) {
                    this.dataMessage.push(cursor.value)
                    cursor = await cursor.continue();
                }
            });
        },
        settingUpWebSocket() {
            if (window["WebSocket"]) {
                this.isWebsocketExist = true;

                let ws = window.location.protocol === 'https:' ? 'wss' : 'ws'
                if (this.myUUID !== "") {
                    this.conn = new WebSocket(ws + "://" + document.location.host + "/ws?name=" + this.myName + "&uuid=" + this.myUUID);
                } else {
                    this.conn = new WebSocket(ws + "://" + document.location.host + "/ws?name=" + this.myName);
                }
                this.conn.onclose = (evt) => {
                    console.log("connection closed")
                };
                this.conn.onmessage = (evt) => {
                    const data = JSON.parse(evt.data)
                    if (data.channel === "chatroom") {
                        if (data.event === "send_message_text") {
                            let message = data.message

                            message.message = this.escapeHtml(message.message)
                            message.message = message.message.replace(/ /g, "&nbsp;");
                            message.message = message.message.replace(/\n/g, "<br/>");

                            dbMessage.set(message.message_id, message).then(() => this.dataMessage.push(message))
                            $(".msg_card_body").stop().animate({scrollTop: $(".msg_card_body")[0].scrollHeight}, 1000);

                        } else if (data.event === "send_message_image") {
                            let message = data.message
                            dbMessage.set(message.message_id, message).then(() => this.dataMessage.push(message))
                            $(".msg_card_body").stop().animate({scrollTop: $(".msg_card_body")[0].scrollHeight}, 1000);
                        } else if (data.event === "unsend_message") {
                            let message = data.message
                            dbMessage.delete(message.message_id).then(() => {
                                this.dataMessage = this.dataMessage.filter(e => e.message_id !== message.message_id)
                            })
                        } else if (data.event === 'status') {
                            let status = data.message
                            if (this.myUUID === '') {
                                let data = {
                                    name: this.myName,
                                    uuid: status.target.uuid,
                                }
                                dbAccount.set(1, data).then(() => this.myUUID = status.target.uuid)
                            }
                            this.toastMessage = status.info
                            $('.toast').toast({
                                autohide: true,
                                delay: 5000,
                            })
                            $('.toast').toast('show')
                            this.onlineTotal = data.message.users.length;
                            this.onlineData = data.message.users;
                        }
                    }

                };
            } else {
                this.isWebsocketExist = false
            }
        },
        newline() {
            // this.textMessage = `${this.textMessage}\r\n`;
        },
        escapeHtml(unsafe) {
            return unsafe
                .replace(/&/g, "&amp;")
                .replace(/</g, "&lt;")
                .replace(/>/g, "&gt;")
                .replace(/"/g, "&quot;")
                .replace(/'/g, "&#039;");
        },
        onPressText(event) {
            const md = new MobileDetect(window.navigator.userAgent);
            if (md.phone() == null && event.keyCode === 13 && !event.shiftKey) {
                this.sendMessage()
            }
        },
        previewImage(data) {
            if (data.file.type === 'image') {
                this.previewImageName = data.file.filename;
                this.previewImageURL = `<img src="${data.file.url}" style="width: 100%" draggable="false">`
                $("#previewImageModal").modal("show")
            }
        },
        async setName() {
            let name = $('#username').val()
            if (name !== "" && name != null) {
                this.myName = name
                await this.settingUpWebSocket()
                setTimeout(() => $("#textMessage").focus(), 100)
            }
        },
        async processUpload(data) {
            const images = data.target.files[0];
            if (images.size > 10000000) {
                this.toastMessage = "Max Uploaded Size Is 10MB"
                $('.toast').toast({
                    autohide: true,
                    delay: 10000,
                })
                $('.toast').toast('show')
            } else {
                const options = {
                    maxSizeMB: 1,
                    maxWidthOrHeight: 600,
                    useWebWorker: true
                }

                try {
                    const compressedFile = await imageCompression(images, options);
                    console.log('compressedFile instanceof Blob', compressedFile instanceof Blob); // true
                    console.log(`compressedFile size ${compressedFile.size / 1024 / 1024} MB`); // smaller than maxSizeMB

                    let formData = new FormData();
                    formData.append("image", compressedFile)

                    const url = "https://api.imgbb.com/1/upload?expiration=10&key=a10724fae77ea714c71642ec9f374890";
                    const option = {
                        method: "POST",
                        body: formData
                    }
                    fetch(url, option)
                        .then(response => response.json())
                        .then(response => this.sendMessageImages(response.data.image));
                } catch (error) {
                    console.log(error);
                }
            }
        },
        sendMessage() {
            if (!(this.conn.readyState === this.conn.OPEN)) {
                return alert("connection closed, please reload")
            }
            if (this.textMessage.trim().length === 0 || this.textMessage == null) {
                return false;
            } else {
                let time = new Date()
                let sendingTime = `${time.getHours()}:${time.getUTCMinutes()}`
                const sending = {
                    name: this.myName,
                    message: this.textMessage,
                    sender: this.myUUID,
                    time: sendingTime,
                    file: {
                        filename: null,
                        type: 'text',
                        url: null
                    }
                };

                this.ws_data.event = "send_message_text"
                this.ws_data.message = sending

                this.conn.send(JSON.stringify(this.ws_data))
                this.textMessage = ""
                $("#textMessage").focus().click()

            }
        },
        sendMessageImages(img) {
            let message = `<img src="${img.url}" draggable="false" style="max-width: 100px; max-height: 100px">`
            let time = new Date()
            let sendingTime = `${time.getHours()}:${time.getUTCMinutes()}`
            const sending = {
                name: this.myName,
                message: message,
                sender: this.myUUID,
                time: sendingTime,
                file: {
                    filename: img.filename,
                    type: 'image',
                    url: img.url
                }
            };

            this.ws_data.event = "send_message_image"
            this.ws_data.message = sending
            this.conn.send(JSON.stringify(this.ws_data))
        },
        unsent(message_id) {
            bootbox.confirm({
                centerVertical: true,
                size: 'small',
                message: "Unsent message ?",
                buttons: {
                    cancel: {
                        label: '<i class="fa fa-times"></i> Cancel',
                        className: 'btn-secondary btn-sm'
                    },
                    confirm: {
                        label: '<i class="fa fa-check"></i> Yes',
                        className: 'btn-danger btn-sm'
                    }
                },
                callback: (result) => {
                    if (result) {
                        this.ws_data.event = "unsend_message"
                        this.ws_data.message = {"message_id": message_id}
                        this.conn.send(JSON.stringify(this.ws_data))
                    }
                }
            });
        },
        async logout() {
            await dbAccount.clear()
            await dbMessage.clear()
            $('.action_menu').toggle();
            this.myName = ""
            this.myUUID = ""
            this.dataMessage = []
            this.conn.close()
        }
    }
})