## Functionality

* 2023/04/23
    1. 先完成了server echo client的request message的功能
* 2023/05/01
    * 假設目前只有一個大廳的情況下
    1. 一個goroutine對應到一個request
    2. 一個額外的goroutine處理broadcast
        1. 出bug啦，怎麼client A發出的消息，自己有收到但client B沒收到啊...
        * 抓到，原來是處理input的goroutine卡住了... 所以需要額外的goroutine去listenToServer...
* 代辦
    1. 多個用戶加入頻道內聊天
        1. client request server 建立頻道
        2. client request server 顯示所有頻道
        3. client request server 加入頻道
        4. ~~client request server 發言，server顯示用戶A的發言給在頻道內的所有人~~ done
    2. 用database儲存頻道聊天訊息

## Debug

1. 開始新的project, 不管怎樣先執行go mod init github.com/your-username/your-project-name 就對惹
