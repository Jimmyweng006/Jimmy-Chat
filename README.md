## Functionality

* 2023/04/23
    1. 先完成了server echo client的request message的功能
* 2023/05/01
    * 假設目前只有一個大廳的情況下
    1. 一個goroutine對應到一個request
    2. 一個額外的goroutine處理broadcast
        1. 出bug啦，怎麼client A發出的消息，自己有收到但client B沒收到啊...
        * 抓到，原來是處理input的goroutine卡住了... 所以需要額外的goroutine去listenToServer...
    * 連接DB
    1. docker run --name jimmy-chat-postgres -e POSTGRES_PASSWORD=root -d -p 5432:5432 postgres
    2. 使用docker compose管理兩個container(Server & DB)
* 2023/05/07
    1. 把之前simple_bank學到的那一套搬過來用囉(golang sqlc, Makefile...)
    2. 完成User登入及發送訊息相關的DB行為
* 2023/05/14
    * 參考了System Design Interview An insider's Guide跟這篇[文章](https://tachunwu.github.io/posts/discord-cassandra/)後，基本上要改成以Cassandra來當作儲存聊天訊息的DB了...
    * Zookeeper: a component to find suitable server(in this case, Chat Server) for clients
* 2023/05/14
    * 先把登入之類的基本功能整合進來吧...
    * 因為沒有設置同時run兩個file出現了funciton undefined問題，搞了好久唉...
        * entrypoint: go run ./server/chat-server.go ./server/api-server.go

* 代辦
    1. 多個用戶加入頻道內聊天
        1. client request server 建立頻道
        2. client request server 顯示所有頻道
        3. client request server 加入頻道
        4. ~~client request server 發言，server顯示用戶A的發言給在頻道內的所有人~~ done
    2. ~~用database儲存頻道聊天訊息~~ done
    3. logIn/signIn
        1. signIn
        2. user password encryption
        3. verify user

## Debug

1. 開始新的project, 不管怎樣先執行go mod init github.com/your-username/your-project-name 就對惹
