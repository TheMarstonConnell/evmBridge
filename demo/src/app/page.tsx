'use client'

import {useAccount, useConnect, useDisconnect, useWriteContract, BaseError} from 'wagmi'
import {useState} from "react";
import {IMerkletreeSource, Merkletree} from "@jackallabs/dogwood-tree";

import './page.css'

function App() {
    const account = useAccount()
    const {connectors, connect, status, connectError} = useConnect()
    const {disconnect} = useDisconnect()
    const {
        data: hash,
        error,
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
            "stateMutability": "payable"
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

    async function doUpload (callback: Function)  {

        const seed = await file.arrayBuffer()
        const source: IMerkletreeSource = {seed: seed, chunkSize: 10240, preserve: false}
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
                callback(root)
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

        console.log("finished")


    }

    const getEthPrice = async (): Promise<number> => {
        try {
            const response = await fetch('https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd');
            const data = await response.json();
            return data.ethereum.usd;
        } catch (error) {
            console.error('Error fetching ETH price:', error);
            throw error;
        }
    };

    function getStoragePrice(price: number, filesize: number): number {

        const storagePrice = 15;     // Price per TB in USD with 8 decimal places
        const multiplier = 2;
        const months = 200 * 12;     // Duration in months (200 years)

        let fs = filesize;
        if (fs <= 1024 * 1024) {
            fs = 1024 * 1024;        // Minimum size is 1 MB
        }

        // Base Storage Multiplier (BSM): storage price calculation
        const BSM = storagePrice * multiplier * months * fs;

        // Calculate price in equivalent "wei" (for parallel logic, use integer math)
        // 1e18 converts to smallest ETH units
        const priceInWei = (BSM * 1e18) / (price * 1099511627776);

        // If the result is zero (due to rounding or too small), set a minimum value
        if (priceInWei === 0) {
            return 5000;  // Return a minimum value (e.g., 5000 units of currency)
        }

        return priceInWei;
    }

    function uploadFile() {
        doUpload((root) => {
            console.log(root)
            getEthPrice().then(price => {
                const p = getStoragePrice(price, file.size);
                writeContract({
                    address: '0x3a5ab5d5df8A8AF40BbcE53DF5999E92b9017483',
                    abi,
                    functionName: 'postFile',
                    args: [root, BigInt(file.size)],
                    value: BigInt(Math.floor(p * 1.01)),
                });
            });

        })

    }

    return (
        <>
            <h1>
                Jackal EVM Demo
            </h1>
            <div id={"account"}>

                <div>
                    {account.status === 'connected' && (
                        <div>
                            <h2>Account</h2>
                            <span>{account.addresses?.[0]}</span>
                            <button className={"discon"} type="button" onClick={() => disconnect()}>
                                Disconnect
                            </button>
                        </div>
                    )}
                </div>


            </div>

            {account.status != 'connected' && (
                <div className={"connectors"}>
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
                    <div>{connectError?.message}</div>
                </div>)}

            <div>
                <h2>Upload</h2>
                <form>
                    <input type="file" onChange={handleFileChange}/>
                </form>
                <button id={"uploadButton"} onClick={uploadFile}
                        disabled={account.status != 'connected' || !file}>Upload
                </button>
                {hash && <div>Transaction Hash: {hash}</div>}
                {isPending && <div>TX Pending...</div>}
                {cid.length > 0 &&
                    <div id={"ipfs"}>IPFS CID: <a target={"_blank"} href={"https://ipfs.io/ipfs/" + cid}>{cid}</a>
                    </div>}
                {hash && cid.length == 0 && <div>File uploading...</div>}
                {error && (
                    <div>Error: {(error as BaseError).shortMessage || error.message}</div>
                )}
            </div>
        </>
    )
}

export default App
