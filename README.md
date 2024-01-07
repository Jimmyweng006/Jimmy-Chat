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
* 2023/07/02
    1. getMessageByRoomID functionality
    2. 開始研究Redis Cache -> 目前應該是採用Write-through的策略，也就是更新DB的時候也更新Cache
    3. 先研究一下全部Cache的效益高不高好了...
* 2023/07/04
    1. Redis應該用來拿最新的100筆資料就好
        1. 測一下200個request的總時長(postgres vs redis)
    2. 應該要有獨立的API讓前端拿chat history，而不是後端從DB拿資料再透過web socket傳出去...
* 2023/12/11
    1. 完全忘記之前7月還沒commit的變更在幹嘛了... 算了準備來研究一下deployment跟CI/CD
        1. [Docker Destkop](https://docs.docker.com/desktop/install/ubuntu/#install-docker-desktop): 三個步驟照著做沒煩惱
        2. ssh key: source的那一方要產生ssh key，然後將public key放到destination server裡。
        3. commands:
            * systemctl --user start docker-desktop
            * sudo systemctl status docker
            * vim /etc/profile
            * sudo systemctl restart docker
* 2023/12/29
    1. ok全部砍掉重弄，乖乖照Digital Ocean那邊的教學走好了...
        1. [How To Install and Use Docker Compose on Ubuntu 22.04](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-docker-compose-on-ubuntu-22-04)
            1. [Initial Server Setup with Ubuntu 22.04](https://www.digitalocean.com/community/tutorials/initial-server-setup-with-ubuntu-22-04)
            2. [How To Install and Use Docker on Ubuntu 22.04](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-docker-on-ubuntu-22-04)
        2. note: 之前沒辦法run docker ps -a/docker-compose up這些指令的問題，在弄完1-1然後用regular user的權限去執行就都好了...
* 2023/12/30
    1. 到logIn之前的步驟都能正常在server上執行了(docker-compose up/make migrateup)
* 2023/12/31
    1. 用Vue來達成client程式的SignIn functionality
* 2024/01/01
    1. 新年快樂！新年第一天繼續開心寫code！在模擬前端API發送到後端的時候，遇到之前常看見但都不是很懂的[CORS (Cross-Origin Resource Sharing)](https://medium.com/starbugs/%E5%BC%84%E6%87%82%E5%90%8C%E6%BA%90%E6%94%BF%E7%AD%96-same-origin-policy-%E8%88%87%E8%B7%A8%E7%B6%B2%E5%9F%9F-cors-e2e5c1a53a19)問題了 -> 只讓特定來源的request可以存取資源，如果不在後端設定的話，即使domain相同但port不同，仍然會被視為不同來源，而被Same Origin Policy擋住而無法存取。
        * 用套件簡單地處理掉以上問題了... 也可以正確寫入註冊的帳號到後端了
    2. 遇到使用者重新整理頁面，再去傳訊息會導致Kafka寫入異常的問題... 好險只是之前close Kafka writer的code亂放導致的...因為重新整理導致Web Socket重新建立，那如果writer也跟著被close，下一次再去寫就整組壞掉(logrus.Fatal)，實際上也不應該去close因為writer是全部人只有一組在共用的。但好險服務還是有被docker restart救回來就是了...
    3. 完成三個流程(Sign In, Login, Chat)的基本功能，準備實驗前端deployment整合！
* 2024/01/03
    1. CI/CD for Backend/Frontend
    2. env url setting for Frontend
* 2024/01/04
    1. env url setting for Backend
    2. local環境下Kafka都能正常收到訊息啊，怎麼上prod就有問題了... -> 好吧container重啟即可...
* 2024/01/07
    1. [setup domain name for digital ocean server](https://docs.digitalocean.com/products/networking/dns/getting-started/dns-registrars/)
    2. 完整的多人對話功能。load chat room message -> TBC


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
    6. ~~CI/CD: 雖然還沒有寫Unit Test所以好像沒有什麼CI可言(?，不過至少推code到Github上後，CD(自動部署)應該要能做到吧！~~
    7. traefik: 好像是新潮的reverse proxy? 之後有空來玩玩
    8. web service: index page
    9. ~~config data for local/prod environment~~

## Learning

### Run Book

1. How to start Backend Service: docker-compose up
2. How to start Frontend Service: docker-compose up --build

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

### Docker

```
sudo systemctl status docker
sudo systemctl restart docker
docker exec -it containerID /bin/sh
```

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
