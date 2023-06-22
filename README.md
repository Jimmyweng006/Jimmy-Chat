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
* 2023/05/21
    * 先把登入之類的基本功能整合進來吧...
    * 因為沒有設置同時run兩個file出現了funciton undefined問題，搞了好久唉...
        * entrypoint: go run ./server/chat-server.go ./server/api-server.go
* 2023/05/23
    * "pq: null value in column \"password\" violates not-null constraint" -> 先從client端開始動工吧...
* 2023/05/29
    * 被"如何在handler裡面處理DB相關的事情"搞得很頭痛，該回去複習之前Clean Architecture的project了...
* 2023/06/03
    1. 先定義好domain的interface(repository, usecase)，再來寫實作...
    2. 依照Clean Architecture的架構，完成client end & server end, User singIn的功能
* 2023/06/04
    1. LogIn功能
    2. 目前預計只有一個公共頻道
    3. 依照User的架構，弄一個Message在Domain Layer裡
    4. 要做到"logIn before chating"好像有點麻煩...
* 2023/06/11
    1. 要做到"logIn before chating"好像有點麻煩... -> 改用JWT的方式來處理驗證身份功能
    2. user開始chatting後會跑出一開始的提示menu -> WaitGroup -> 乖乖block main goroutine就好啦...
    3. ~~除了GetByUsername(logIn的時候會用到)，還要GetByUserID，這樣才能跟後面的createMessage整合~~ -> ID的資訊其實只要把client object傳進來就能拿到了...
    4. ~~ID的資訊其實只要把client object傳進來就能拿到了...~~ -> ID的資訊應該在message struct裡... 也就是該這樣寫 senderID, err := strconv.Atoi(messageObject.Sender)
    5. 把kafka的服務整合到，傳訊息給聊天室其他user的code裡面！
        1. brokers 要使用 []string{"kafka:9092"}, 而不是 []string{"localhost:9092"}...
        2. 怎麼會有還沒write就能read的奇怪問題發生... "Received message from Kafka: \"MTIz\"\n"
        3. 奇怪的字的問題應該是webSocket讀到的資料([]byte)再去做Marshal([]byte)導致的...
* 2023/06/12
    1. 好不容易把kafka-ui也整合進docker了，結果看不到資料？？？
    2. kafka資料會重複讀 -> multiple reader problem
* 2023/06/22
    1. Kafka knowledge
    2. get user data by user id function

* 代辦
    1. 多個用戶加入頻道內聊天
        1. client request server 建立頻道
        2. client request server 顯示所有頻道
        3. client request server 加入頻道
        4. ~~client request server 發言，server顯示用戶A的發言給在頻道內的所有人~~ done
    2. ~~用database儲存頻道聊天訊息~~ done
    3. logIn/signIn
        1. ~~signIn~~
        2. ~~user password encryption~~
        3. ~~verify user~~
        4. ~~logIn before chating!!!~~
    4. 用Kafka減緩Chat Server傳送大量訊息的壓力
    5. 拿到聊天室的聊天資訊
        1. 比較從Redis拿資料跟直接從Postgres拿資料的效能差異

## Learning

### Kafka

1. Top-Down Hierarchy
    1. Cluster: 多個Broker組成Cluster
    2. Broker: 類似Server的概念
    3. Topic:
    4. Partition:
    5. Record: {Key, Value, TimeStamp}
    6. Batch: 多筆Record成為一個Batch，再寫入Kafka。

## Debug

1. 開始新的project, 不管怎樣先執行go mod init github.com/your-username/your-project-name 就對惹

## Reference

1. [DB table design](https://dbdiagram.io/d/644fb728dca9fb07c44eff8b)
2. [Kafka](https://ithelp.ithome.com.tw/users/20140255/ironman/4026?page=1)
