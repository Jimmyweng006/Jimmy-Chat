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
        1. brokers 要使用 []string{"kafka:9092"}, 而不是 []string{"localhost:9092"}... -> broker:29091 -> KAFKA_BOOTSTRAP.SERVERS
        2. 怎麼會有還沒write就能read的奇怪問題發生... "Received message from Kafka: \"MTIz\"\n"
        3. 奇怪的字的問題應該是webSocket讀到的資料([]byte)再去做Marshal([]byte)導致的...
* 2023/06/12
    1. 好不容易把kafka-ui也整合進docker了，結果看不到資料？？？ -> 簡單用個網路的範例confluentinc/cp-enterprise-control-center就好
    2. kafka資料會重複讀 -> multiple reader problem
* 2023/06/22
    1. Kafka knowledge
    2. get user data by user id function
* 2023/06/23
    1. Kafka knowledge
    2. multiple reader problem -> 直接改成 1 reader & 1 writer後目前看來沒有奇怪的Bug，但效能問題的話要再壓測一下...
    3. no required module provides package "github.com/Jimmyweng006/Jimmy-Chat/messageQueue -> 一直出現的奇怪問題？不知道是不是改成小寫或是重開VS Code，問題就解決了...
    4. MessageQueueWrapper設計思考
        1. 因為希望NewKafka()回傳的是Interface(如果之後要換不同的MQ比較方便)
        2. 而且回傳的是指標類型(不用複製整個Struct效能應該好一些)
        3. 不再多用type MessageQueueWrapper struct包一層的話，回來的資料型態會是指標介面，這樣的話無法使用介面的方法...
        4. 所以參考db.Queries的設計弄出來了，下面是對應關係
            1. type Queries struct <-> type MessageQueueWrapper struct
            2. type DBTX interface <-> type MessageQueue interface
* 2023/06/24
    1. 先來準備壓力測試的工具
        1. 折騰一番終於設定好Jmeter了... path for pepper-box-1.0.jar: /opt/homebrew/Cellar/jmeter/5.5/libexec/lib/ext
        2. 結果local的Jmeter一直連不到docker內的Kafka...算了土法煉鋼直接hard code測試...
* 2023/06/25
    1. load testing: how much time it takes to perform 1000 write: 16m46.240430001s -> QPS約是1，厲害了...
    2. 無法控制使用者寫的速度，那就先來改善reader的效率吧... -> 等等寫也很慢啊...
    3. 重新啟動server後，在Kafka的舊資料又會被重新讀一次 -> 應該要更新offset??? -> ReadMessage automatically commits offsets when using consumer groups.
    4. 要enhance成多個reader的話在現有架構要怎麼改呢...
* 2023/06/26
    1. enhance writer(writeMessage -> writeMessages)功能後，怎麼client一使用/chat就斷線... -> panic: runtime error: index out of range [201] with length 201，還敢忘記MOD啊....
    2. one writer & 3 reader(?), 處理1000筆資料的時間大約是5s左右... -> 之後嘗試更多筆資料(1w以上)看看！

* 代辦
    1. 多個用戶加入頻道內聊天
        1. client request server 建立頻道
        2. client request server 顯示所有頻道
        3. client request server 加入頻道
        4. ~~client request server 發言，server顯示用戶A的發言給在頻道內的所有人~~ done
    2. ~~用database儲存頻道聊天訊息~~ done
    3. logIn/signIn
        1. ~~signIn~~ done
        2. ~~user password encryption~~ done
        3. ~~verify user~~ done
        4. ~~logIn before chating!!!~~ done
    4. 用Kafka減緩Chat Server傳送大量訊息的壓力
        1. 比較multiple reader/writer 跟 1 reader/writer的效能差異
    5. 拿到聊天室的聊天資訊
        1. 比較從Redis拿資料跟直接從Postgres拿資料的效能差異

## Learning

### Kafka

1. Terminology(Top-Down Hierarchy)
    1. Cluster: 多個Broker組成Cluster
    2. Broker: 類似Server的概念
    3. Topic: 類似DB的Table，但是只會一直加資料不會刪除資料。
    4. Partition: Topic的所有資料再切分成一或多個區段，一個資料只會去到一個區段！
    5. Offset: 下一次Consume的起始點
    6. Replication: 同一筆資料存在幾台Broker上

## Debug

1. 開始新的project, 不管怎樣先執行go mod init github.com/your-username/your-project-name 就對惹

### Postgres

```
Identify what is running in port 5432: sudo lsof -i :5432

Kill all the processes that are running under this port: sudo kill -9 <pid>

Run the command again to verify no process is running now: sudo lsof -i :5432
```

## Reference

1. [DB table design](https://dbdiagram.io/d/644fb728dca9fb07c44eff8b)
2. Kafka
    1. [introduce](https://chrisyen8341.medium.com/kafka%E8%B6%85%E6%96%B0%E6%89%8B%E5%85%A5%E9%96%80%E7%AC%AC%E4%B8%80%E7%9E%A5-9348a9cb23dc)
    2. [setup](https://morosedog.gitlab.io/docker-20201116-docker17/)
