'use client'

import {useAccount, useConnect, useDisconnect, useWriteContract} from 'wagmi'
import {useState} from "react";
import {IMerkletreeSource, Merkletree} from "@jackallabs/dogwood-tree";

function App() {
  const account = useAccount()
  const { connectors, connect, status, error } = useConnect()
  const { disconnect } = useDisconnect()
    const {
        data: hash,
        isPending,
        writeContract
    } = useWriteContract()

    const [file, setFile] = useState(null);
    const [cid, setCid] = useState("");

    const handleFileChange = (event) => {
        const selectedFile = event.target.files[0];
        if (selectedFile) {
            setFile(selectedFile);
        }
    };


    const abi = [
        {
            "type": "function",
            "name": "postFile",
            "inputs": [
                {
                    "name": "merkle",
                    "type": "string",
                    "internalType": "string"
                },
                {
                    "name": "filesize",
                    "type": "uint64",
                    "internalType": "uint64"
                }
            ],
            "outputs": [],
            "stateMutability": "nonpayable"
        },
        {
            "type": "event",
            "name": "PostedFile",
            "inputs": [
                {
                    "name": "sender",
                    "type": "address",
                    "indexed": false,
                    "internalType": "address"
                },
                {
                    "name": "merkle",
                    "type": "string",
                    "indexed": false,
                    "internalType": "string"
                },
                {
                    "name": "size",
                    "type": "uint64",
                    "indexed": false,
                    "internalType": "uint64"
                }
            ],
            "anonymous": false
        }
    ]

    const uploadFile = async () => {

        const seed = await file.arrayBuffer()
        const source: IMerkletreeSource = { seed: seed, chunkSize: 10240, preserve: false }
        const tree = await Merkletree.grow(source)
        const root = tree.getRootAsHex()

        console.log(`Root: ${root}`)

        // Replace with your Tendermint WebSocket endpoint
        const wsUrl = 'wss://testnet-rpc.jackalprotocol.com/websocket';

        // Create a new WebSocket connection
        const socket = new WebSocket(wsUrl);

        // Connection opened
        socket.addEventListener("open", (event) => {
            console.log('Connected to the WebSocket');

            // Subscribe to a specific event (e.g., new block header)
            const subscriptionMessage = JSON.stringify({
                "jsonrpc": "2.0",
                "method": "subscribe",
                "id": "1",
                "params": {
                    "query": `tm.event='Tx' AND post_file.file='${root}'`
                }
            });

            socket.send(subscriptionMessage);
            setTimeout(() => {
                socket.close()
            }, 60000)

        });

        socket.addEventListener("message", async (event) => {
            const data = JSON.parse(event.data)
            console.log(data)
            console.log(data.result)
            if (Object.keys(data.result).length == 0) {
                return
            }

            const startS = data.result.events["post_file.start"][0]
            const senderS = data.result.events["post_file.signer"][0]

            const url = 'https://testnet-provider.jackallabs.io/upload';

            // Create a FormData object
            const formData = new FormData();

            // Append the file to the form data
            // Replace this with an actual File object if using in a browser
            formData.append('file', file);
            // Append the sender and merkle fields
            formData.append('sender', senderS);
            formData.append('merkle', root);
            formData.append('start', Number(startS));



            const request = new Request(url, {
                method: "POST",
                body: formData,
            });


            try {
                // Send the POST request using fetch
                const response = await fetch(request);

                // Handle the response
                if (!response.ok) {
                    throw new Error(`Upload failed with status: ${response.status}`);
                }

                const data = await response.json();

                const cid = data["cid"]
                setCid(cid)

            } catch (error) {
                console.error('Upload failed:', error);
            }


            socket.close()
        });

        writeContract({
            address: '0x5FbDB2315678afecb367f032d93F642f64180aa3',
            abi,
            functionName: 'postFile',
            args: [root, BigInt(file.size)],
        })




    }
  return (
      <>
        <div>
          <h2>Account</h2>

          <div>
            status: {account.status}
            <br/>
            addresses: {JSON.stringify(account.addresses)}
            <br/>
            chainId: {account.chainId}
          </div>

          {account.status === 'connected' && (
              <button type="button" onClick={() => disconnect()}>
                Disconnect
              </button>
          )}
        </div>

        <div>
          <h2>Connect</h2>
          {connectors.map((connector) => (
              <button
                  key={connector.uid}
                  onClick={() => connect({connector})}
                  type="button"
              >
                {connector.name}
              </button>
          ))}
          <div>{status}</div>
          <div>{error?.message}</div>
        </div>

          <div>
              <h2>Upload</h2>
              <form>
                  <input type="file" onChange={handleFileChange}/>
              </form>
              <button onClick={uploadFile}>Upload</button>
              {hash && <div>Transaction Hash: {hash}</div>}
              {cid.length > 0 && <div>IPFS CID: {cid}</div>}

          </div>
      </>
  )
}

export default App
