# nchc-vmbr


## VMBR（VM Backup & Restore）

`vmbr` 代表 VM Backup & Restore（虛擬機備份與還原）。保存 VM 的狀態，並在需要時從備份重建或還原 VM。

- **備份流程（Backup Flow）**：
    - **Snapshot**：對 VM 建立快照以捕捉當前狀態與磁碟映像。
    - **Download**：快照或映像存入VRM的Repository，以 Tag 管理版本之後，下載至CS的物件儲存
    - **Transfer**：透過rclone在不同的物件儲存間備份（例如從 CS 到 Shared S3）。

- **還原流程（Restore Flow）**：
    - **Transfer**：透過rcloen從Shared S3 取回映像檔至CS。
    - **Upload**：從CS將映像檔上傳至VRM的Repository，以 Tag 管理版本。
    - **Create VM**：使用VRM的映像檔還原之前建立快照的VM。

## VM Backup & Restore Flow

```
    Backup Flow                        Restore Flow  
------------------                -------------------
      +---------+                       +---------+
      | Backup  |                       | Restore |
      |  VM     |                       |   VM    |
      +---------+                       +---------+
           |                                 ^
           | 1. Snapshot                     | 6. Create 
           v                                 |
      +---------+                       +---------+
      |  VRM    |                       |   VRM   |
      | Repo:Tag|                       | Repo:Tag|
      +---------+                       +---------+
           |                                 ^
           | 2. Download                     | 5. Upload
           v                                 |
       +-------+                        +-------+
       |  CS   |                        |   CS  |
       +-------+                        +-------+
           |                                 ^
           | 3. Transfer                     | 4. Transfer
           v                                 |
   +------------------+               +------------------+
   |   Shared S3      |<------------->|   Shared S3      |
   +------------------+               +------------------+

```




