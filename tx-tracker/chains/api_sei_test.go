package chains

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const jsonTxSearchResponse = `
{
	"jsonrpc": "2.0",
	"id": 1729346572401935400,
	"result": {
		"txs": [
			{
				"hash": "D97FD8EB0FAB7784A8A293A7FEF1F47FDE0C4375C254A19361E0F87CC01EF99A",
				"height": "33472147",
				"index": 34,
				"tx_result": {
					"data": "CiYKJC9jb3Ntd2FzbS53YXNtLnYxLk1zZ0V4ZWN1dGVDb250cmFjdA==",
					"log": "[{\"events\":[{\"type\":\"burn\",\"attributes\":[{\"key\":\"burner\",\"value\":\"sei19ejy8n9qsectrf4semdp9cpknflld0j6svvmtq\"},{\"key\":\"amount\",\"value\":\"1000factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/3ApLjovgkMT4LWAcqyYDPaNiDSKmuJJfMom18Ed29o27\"},{\"key\":\"burn_from_address\",\"value\":\"sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l\"},{\"key\":\"amount\",\"value\":\"1000factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/3ApLjovgkMT4LWAcqyYDPaNiDSKmuJJfMom18Ed29o27\"}]},{\"type\":\"coin_received\",\"attributes\":[{\"key\":\"receiver\",\"value\":\"sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l\"},{\"key\":\"amount\",\"value\":\"1000factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/3ApLjovgkMT4LWAcqyYDPaNiDSKmuJJfMom18Ed29o27\"},{\"key\":\"receiver\",\"value\":\"sei19ejy8n9qsectrf4semdp9cpknflld0j6svvmtq\"},{\"key\":\"amount\",\"value\":\"1000factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/3ApLjovgkMT4LWAcqyYDPaNiDSKmuJJfMom18Ed29o27\"}]},{\"type\":\"coin_spent\",\"attributes\":[{\"key\":\"spender\",\"value\":\"sei17dxuvdfgxu0gpym3hu8glcct9kjccn4xtdfgfc\"},{\"key\":\"amount\",\"value\":\"1000factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/3ApLjovgkMT4LWAcqyYDPaNiDSKmuJJfMom18Ed29o27\"},{\"key\":\"spender\",\"value\":\"sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l\"},{\"key\":\"amount\",\"value\":\"1000factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/3ApLjovgkMT4LWAcqyYDPaNiDSKmuJJfMom18Ed29o27\"},{\"key\":\"spender\",\"value\":\"sei19ejy8n9qsectrf4semdp9cpknflld0j6svvmtq\"},{\"key\":\"amount\",\"value\":\"1000factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/3ApLjovgkMT4LWAcqyYDPaNiDSKmuJJfMom18Ed29o27\"}]},{\"type\":\"execute\",\"attributes\":[{\"key\":\"_contract_address\",\"value\":\"sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l\"},{\"key\":\"_contract_address\",\"value\":\"sei1yqajzpwm4ud53jkhcndy576p6tfpp3sjecrg6keurm3l46kj6pyq5p2mhw\"},{\"key\":\"_contract_address\",\"value\":\"sei1smzlm9t79kur392nu9egl8p8je9j92q4gzguewj56a05kyxxra0qy0nuf3\"},{\"key\":\"_contract_address\",\"value\":\"sei1yqajzpwm4ud53jkhcndy576p6tfpp3sjecrg6keurm3l46kj6pyq5p2mhw\"},{\"key\":\"_contract_address\",\"value\":\"sei1gjrrme22cyha4ht2xapn3f08zzw6z3d4uxx6fyy9zd5dyr3yxgzqqncdqn\"}]},{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"/cosmwasm.wasm.v1.MsgExecuteContract\"},{\"key\":\"module\",\"value\":\"wasm\"},{\"key\":\"sender\",\"value\":\"sei17dxuvdfgxu0gpym3hu8glcct9kjccn4xtdfgfc\"}]},{\"type\":\"send_packet\",\"attributes\":[{\"key\":\"packet_data\",\"value\":\"{\\\"publish\\\":{\\\"msg\\\":[{\\\"key\\\":\\\"message.message\\\",\\\"value\\\":\\\"0100000000000000000000000000000000000000000000000000000000000003e8069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f000000000010001efe18e2a3342366d5d0823766989514c907a243667dd9ff2a4c3fc46d28ca23f00010000000000000000000000000000000000000000000000000000000000000000\\\"},{\\\"key\\\":\\\"message.sender\\\",\\\"value\\\":\\\"86c5fd957e2db8389553e1728f9c27964b22a8154091ccba54d75f4b10c61f5e\\\"},{\\\"key\\\":\\\"message.chain_id\\\",\\\"value\\\":\\\"32\\\"},{\\\"key\\\":\\\"message.nonce\\\",\\\"value\\\":\\\"0\\\"},{\\\"key\\\":\\\"message.sequence\\\",\\\"value\\\":\\\"15557\\\"},{\\\"key\\\":\\\"message.block_time\\\",\\\"value\\\":\\\"1697810572\\\"}]}}\"},{\"key\":\"packet_data_hex\",\"value\":\"7b227075626c697368223a7b226d7367223a5b7b226b6579223a226d6573736167652e6d657373616765222c2276616c7565223a223031303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303365383036396238383537666561623831383466623638376636333436313863303335646163343339646331616562336235353938613066303030303030303030303130303031656665313865326133333432333636643564303832333736363938393531346339303761323433363637646439666632613463336663343664323863613233663030303130303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030227d2c7b226b6579223a226d6573736167652e73656e646572222c2276616c7565223a2238366335666439353765326462383338393535336531373238663963323739363462323261383135343039316363626135346437356634623130633631663565227d2c7b226b6579223a226d6573736167652e636861696e5f6964222c2276616c7565223a223332227d2c7b226b6579223a226d6573736167652e6e6f6e6365222c2276616c7565223a2230227d2c7b226b6579223a226d6573736167652e73657175656e6365222c2276616c7565223a223135353537227d2c7b226b6579223a226d6573736167652e626c6f636b5f74696d65222c2276616c7565223a2231363937383130353732227d5d7d7d\"},{\"key\":\"packet_timeout_height\",\"value\":\"0-0\"},{\"key\":\"packet_timeout_timestamp\",\"value\":\"1729346572401935236\"},{\"key\":\"packet_sequence\",\"value\":\"15562\"},{\"key\":\"packet_src_port\",\"value\":\"wasm.sei1gjrrme22cyha4ht2xapn3f08zzw6z3d4uxx6fyy9zd5dyr3yxgzqqncdqn\"},{\"key\":\"packet_src_channel\",\"value\":\"channel-4\"},{\"key\":\"packet_dst_port\",\"value\":\"wasm.wormhole1wkwy0xh89ksdgj9hr347dyd2dw7zesmtrue6kfzyml4vdtz6e5ws2y050r\"},{\"key\":\"packet_dst_channel\",\"value\":\"channel-0\"},{\"key\":\"packet_channel_ordering\",\"value\":\"ORDER_UNORDERED\"},{\"key\":\"packet_connection\",\"value\":\"connection-6\"}]},{\"type\":\"transfer\",\"attributes\":[{\"key\":\"recipient\",\"value\":\"sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l\"},{\"key\":\"sender\",\"value\":\"sei17dxuvdfgxu0gpym3hu8glcct9kjccn4xtdfgfc\"},{\"key\":\"amount\",\"value\":\"1000factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/3ApLjovgkMT4LWAcqyYDPaNiDSKmuJJfMom18Ed29o27\"},{\"key\":\"recipient\",\"value\":\"sei19ejy8n9qsectrf4semdp9cpknflld0j6svvmtq\"},{\"key\":\"sender\",\"value\":\"sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l\"},{\"key\":\"amount\",\"value\":\"1000factory/sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l/3ApLjovgkMT4LWAcqyYDPaNiDSKmuJJfMom18Ed29o27\"}]},{\"type\":\"wasm\",\"attributes\":[{\"key\":\"_contract_address\",\"value\":\"sei1yqajzpwm4ud53jkhcndy576p6tfpp3sjecrg6keurm3l46kj6pyq5p2mhw\"},{\"key\":\"action\",\"value\":\"increase_allowance\"},{\"key\":\"owner\",\"value\":\"sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l\"},{\"key\":\"spender\",\"value\":\"sei1smzlm9t79kur392nu9egl8p8je9j92q4gzguewj56a05kyxxra0qy0nuf3\"},{\"key\":\"amount\",\"value\":\"1000\"},{\"key\":\"_contract_address\",\"value\":\"sei1smzlm9t79kur392nu9egl8p8je9j92q4gzguewj56a05kyxxra0qy0nuf3\"},{\"key\":\"transfer.token_chain\",\"value\":\"1\"},{\"key\":\"transfer.token\",\"value\":\"069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f00000000001\"},{\"key\":\"transfer.sender\",\"value\":\"397ad473aee22d1cd2829ea2238cbffa8a61c9d8eea3a1f7ccf20dc14ca78188\"},{\"key\":\"transfer.recipient_chain\",\"value\":\"1\"},{\"key\":\"transfer.recipient\",\"value\":\"efe18e2a3342366d5d0823766989514c907a243667dd9ff2a4c3fc46d28ca23f\"},{\"key\":\"transfer.amount\",\"value\":\"1000\"},{\"key\":\"transfer.nonce\",\"value\":\"0\"},{\"key\":\"transfer.block_time\",\"value\":\"1697810572\"},{\"key\":\"_contract_address\",\"value\":\"sei1yqajzpwm4ud53jkhcndy576p6tfpp3sjecrg6keurm3l46kj6pyq5p2mhw\"},{\"key\":\"action\",\"value\":\"burn_from\"},{\"key\":\"from\",\"value\":\"sei189adguawugk3e55zn63z8r9ll29xrjwca636ra7v7gxuzn98sxyqwzt47l\"},{\"key\":\"by\",\"value\":\"sei1smzlm9t79kur392nu9egl8p8je9j92q4gzguewj56a05kyxxra0qy0nuf3\"},{\"key\":\"amount\",\"value\":\"1000\"},{\"key\":\"_contract_address\",\"value\":\"sei1gjrrme22cyha4ht2xapn3f08zzw6z3d4uxx6fyy9zd5dyr3yxgzqqncdqn\"},{\"key\":\"message.message\",\"value\":\"0100000000000000000000000000000000000000000000000000000000000003e8069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f000000000010001efe18e2a3342366d5d0823766989514c907a243667dd9ff2a4c3fc46d28ca23f00010000000000000000000000000000000000000000000000000000000000000000\"},{\"key\":\"message.sender\",\"value\":\"86c5fd957e2db8389553e1728f9c27964b22a8154091ccba54d75f4b10c61f5e\"},{\"key\":\"message.chain_id\",\"value\":\"32\"},{\"key\":\"message.nonce\",\"value\":\"0\"},{\"key\":\"message.sequence\",\"value\":\"15557\"},{\"key\":\"message.block_time\",\"value\":\"1697810572\"},{\"key\":\"is_ibc\",\"value\":\"true\"}]}]}]",
					"gas_wanted": "903925",
					"gas_used": "644246",
					"events": [
						{
							"type": "coin_spent",
							"attributes": [
								{
									"key": "c3BlbmRlcg==",
									"value": "c2VpMTdkeHV2ZGZneHUwZ3B5bTNodThnbGNjdDlramNjbjR4dGRmZ2Zj",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "OTAzOTN1c2Vp",
									"index": true
								}
							]
						},
						{
							"type": "tx",
							"attributes": [
								{
									"key": "ZmVl",
									"value": "OTAzOTN1c2Vp",
									"index": true
								},
								{
									"key": "ZmVlX3BheWVy",
									"value": "c2VpMTdkeHV2ZGZneHUwZ3B5bTNodThnbGNjdDlramNjbjR4dGRmZ2Zj",
									"index": true
								}
							]
						},
						{
							"type": "tx",
							"attributes": [
								{
									"key": "YWNjX3NlcQ==",
									"value": "c2VpMTdkeHV2ZGZneHUwZ3B5bTNodThnbGNjdDlramNjbjR4dGRmZ2ZjLzU3",
									"index": true
								}
							]
						},
						{
							"type": "tx",
							"attributes": [
								{
									"key": "c2lnbmF0dXJl",
									"value": "UjBzNWw2NWNvbE9pU1l4c1FMRU10R1craVkzSENROEdvRFFtd2tVNStNVVNnQjVXaWZORThnYzlnQW5uZThQaTdETk5KdVpZTlpXTkFld1NUMjRremc9PQ==",
									"index": true
								}
							]
						},
						{
							"type": "message",
							"attributes": [
								{
									"key": "YWN0aW9u",
									"value": "L2Nvc213YXNtLndhc20udjEuTXNnRXhlY3V0ZUNvbnRyYWN0",
									"index": true
								}
							]
						},
						{
							"type": "message",
							"attributes": [
								{
									"key": "bW9kdWxl",
									"value": "d2FzbQ==",
									"index": true
								},
								{
									"key": "c2VuZGVy",
									"value": "c2VpMTdkeHV2ZGZneHUwZ3B5bTNodThnbGNjdDlramNjbjR4dGRmZ2Zj",
									"index": true
								}
							]
						},
						{
							"type": "coin_spent",
							"attributes": [
								{
									"key": "c3BlbmRlcg==",
									"value": "c2VpMTdkeHV2ZGZneHUwZ3B5bTNodThnbGNjdDlramNjbjR4dGRmZ2Zj",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "MTAwMGZhY3Rvcnkvc2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2wvM0FwTGpvdmdrTVQ0TFdBY3F5WURQYU5pRFNLbXVKSmZNb20xOEVkMjlvMjc=",
									"index": true
								}
							]
						},
						{
							"type": "coin_received",
							"attributes": [
								{
									"key": "cmVjZWl2ZXI=",
									"value": "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "MTAwMGZhY3Rvcnkvc2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2wvM0FwTGpvdmdrTVQ0TFdBY3F5WURQYU5pRFNLbXVKSmZNb20xOEVkMjlvMjc=",
									"index": true
								}
							]
						},
						{
							"type": "transfer",
							"attributes": [
								{
									"key": "cmVjaXBpZW50",
									"value": "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
									"index": true
								},
								{
									"key": "c2VuZGVy",
									"value": "c2VpMTdkeHV2ZGZneHUwZ3B5bTNodThnbGNjdDlramNjbjR4dGRmZ2Zj",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "MTAwMGZhY3Rvcnkvc2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2wvM0FwTGpvdmdrTVQ0TFdBY3F5WURQYU5pRFNLbXVKSmZNb20xOEVkMjlvMjc=",
									"index": true
								}
							]
						},
						{
							"type": "execute",
							"attributes": [
								{
									"key": "X2NvbnRyYWN0X2FkZHJlc3M=",
									"value": "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
									"index": true
								}
							]
						},
						{
							"type": "coin_spent",
							"attributes": [
								{
									"key": "c3BlbmRlcg==",
									"value": "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "MTAwMGZhY3Rvcnkvc2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2wvM0FwTGpvdmdrTVQ0TFdBY3F5WURQYU5pRFNLbXVKSmZNb20xOEVkMjlvMjc=",
									"index": true
								}
							]
						},
						{
							"type": "coin_received",
							"attributes": [
								{
									"key": "cmVjZWl2ZXI=",
									"value": "c2VpMTllank4bjlxc2VjdHJmNHNlbWRwOWNwa25mbGxkMGo2c3Z2bXRx",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "MTAwMGZhY3Rvcnkvc2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2wvM0FwTGpvdmdrTVQ0TFdBY3F5WURQYU5pRFNLbXVKSmZNb20xOEVkMjlvMjc=",
									"index": true
								}
							]
						},
						{
							"type": "transfer",
							"attributes": [
								{
									"key": "cmVjaXBpZW50",
									"value": "c2VpMTllank4bjlxc2VjdHJmNHNlbWRwOWNwa25mbGxkMGo2c3Z2bXRx",
									"index": true
								},
								{
									"key": "c2VuZGVy",
									"value": "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "MTAwMGZhY3Rvcnkvc2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2wvM0FwTGpvdmdrTVQ0TFdBY3F5WURQYU5pRFNLbXVKSmZNb20xOEVkMjlvMjc=",
									"index": true
								}
							]
						},
						{
							"type": "coin_spent",
							"attributes": [
								{
									"key": "c3BlbmRlcg==",
									"value": "c2VpMTllank4bjlxc2VjdHJmNHNlbWRwOWNwa25mbGxkMGo2c3Z2bXRx",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "MTAwMGZhY3Rvcnkvc2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2wvM0FwTGpvdmdrTVQ0TFdBY3F5WURQYU5pRFNLbXVKSmZNb20xOEVkMjlvMjc=",
									"index": true
								}
							]
						},
						{
							"type": "burn",
							"attributes": [
								{
									"key": "YnVybmVy",
									"value": "c2VpMTllank4bjlxc2VjdHJmNHNlbWRwOWNwa25mbGxkMGo2c3Z2bXRx",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "MTAwMGZhY3Rvcnkvc2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2wvM0FwTGpvdmdrTVQ0TFdBY3F5WURQYU5pRFNLbXVKSmZNb20xOEVkMjlvMjc=",
									"index": true
								}
							]
						},
						{
							"type": "burn",
							"attributes": [
								{
									"key": "YnVybl9mcm9tX2FkZHJlc3M=",
									"value": "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "MTAwMGZhY3Rvcnkvc2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2wvM0FwTGpvdmdrTVQ0TFdBY3F5WURQYU5pRFNLbXVKSmZNb20xOEVkMjlvMjc=",
									"index": true
								}
							]
						},
						{
							"type": "execute",
							"attributes": [
								{
									"key": "X2NvbnRyYWN0X2FkZHJlc3M=",
									"value": "c2VpMXlxYWp6cHdtNHVkNTNqa2hjbmR5NTc2cDZ0ZnBwM3NqZWNyZzZrZXVybTNsNDZrajZweXE1cDJtaHc=",
									"index": true
								}
							]
						},
						{
							"type": "wasm",
							"attributes": [
								{
									"key": "X2NvbnRyYWN0X2FkZHJlc3M=",
									"value": "c2VpMXlxYWp6cHdtNHVkNTNqa2hjbmR5NTc2cDZ0ZnBwM3NqZWNyZzZrZXVybTNsNDZrajZweXE1cDJtaHc=",
									"index": true
								},
								{
									"key": "YWN0aW9u",
									"value": "aW5jcmVhc2VfYWxsb3dhbmNl",
									"index": true
								},
								{
									"key": "b3duZXI=",
									"value": "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
									"index": true
								},
								{
									"key": "c3BlbmRlcg==",
									"value": "c2VpMXNtemxtOXQ3OWt1cjM5Mm51OWVnbDhwOGplOWo5MnE0Z3pndWV3ajU2YTA1a3l4eHJhMHF5MG51ZjM=",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "MTAwMA==",
									"index": true
								}
							]
						},
						{
							"type": "execute",
							"attributes": [
								{
									"key": "X2NvbnRyYWN0X2FkZHJlc3M=",
									"value": "c2VpMXNtemxtOXQ3OWt1cjM5Mm51OWVnbDhwOGplOWo5MnE0Z3pndWV3ajU2YTA1a3l4eHJhMHF5MG51ZjM=",
									"index": true
								}
							]
						},
						{
							"type": "wasm",
							"attributes": [
								{
									"key": "X2NvbnRyYWN0X2FkZHJlc3M=",
									"value": "c2VpMXNtemxtOXQ3OWt1cjM5Mm51OWVnbDhwOGplOWo5MnE0Z3pndWV3ajU2YTA1a3l4eHJhMHF5MG51ZjM=",
									"index": true
								},
								{
									"key": "dHJhbnNmZXIudG9rZW5fY2hhaW4=",
									"value": "MQ==",
									"index": true
								},
								{
									"key": "dHJhbnNmZXIudG9rZW4=",
									"value": "MDY5Yjg4NTdmZWFiODE4NGZiNjg3ZjYzNDYxOGMwMzVkYWM0MzlkYzFhZWIzYjU1OThhMGYwMDAwMDAwMDAwMQ==",
									"index": true
								},
								{
									"key": "dHJhbnNmZXIuc2VuZGVy",
									"value": "Mzk3YWQ0NzNhZWUyMmQxY2QyODI5ZWEyMjM4Y2JmZmE4YTYxYzlkOGVlYTNhMWY3Y2NmMjBkYzE0Y2E3ODE4OA==",
									"index": true
								},
								{
									"key": "dHJhbnNmZXIucmVjaXBpZW50X2NoYWlu",
									"value": "MQ==",
									"index": true
								},
								{
									"key": "dHJhbnNmZXIucmVjaXBpZW50",
									"value": "ZWZlMThlMmEzMzQyMzY2ZDVkMDgyMzc2Njk4OTUxNGM5MDdhMjQzNjY3ZGQ5ZmYyYTRjM2ZjNDZkMjhjYTIzZg==",
									"index": true
								},
								{
									"key": "dHJhbnNmZXIuYW1vdW50",
									"value": "MTAwMA==",
									"index": true
								},
								{
									"key": "dHJhbnNmZXIubm9uY2U=",
									"value": "MA==",
									"index": true
								},
								{
									"key": "dHJhbnNmZXIuYmxvY2tfdGltZQ==",
									"value": "MTY5NzgxMDU3Mg==",
									"index": true
								}
							]
						},
						{
							"type": "execute",
							"attributes": [
								{
									"key": "X2NvbnRyYWN0X2FkZHJlc3M=",
									"value": "c2VpMXlxYWp6cHdtNHVkNTNqa2hjbmR5NTc2cDZ0ZnBwM3NqZWNyZzZrZXVybTNsNDZrajZweXE1cDJtaHc=",
									"index": true
								}
							]
						},
						{
							"type": "wasm",
							"attributes": [
								{
									"key": "X2NvbnRyYWN0X2FkZHJlc3M=",
									"value": "c2VpMXlxYWp6cHdtNHVkNTNqa2hjbmR5NTc2cDZ0ZnBwM3NqZWNyZzZrZXVybTNsNDZrajZweXE1cDJtaHc=",
									"index": true
								},
								{
									"key": "YWN0aW9u",
									"value": "YnVybl9mcm9t",
									"index": true
								},
								{
									"key": "ZnJvbQ==",
									"value": "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
									"index": true
								},
								{
									"key": "Ynk=",
									"value": "c2VpMXNtemxtOXQ3OWt1cjM5Mm51OWVnbDhwOGplOWo5MnE0Z3pndWV3ajU2YTA1a3l4eHJhMHF5MG51ZjM=",
									"index": true
								},
								{
									"key": "YW1vdW50",
									"value": "MTAwMA==",
									"index": true
								}
							]
						},
						{
							"type": "execute",
							"attributes": [
								{
									"key": "X2NvbnRyYWN0X2FkZHJlc3M=",
									"value": "c2VpMWdqcnJtZTIyY3loYTRodDJ4YXBuM2YwOHp6dzZ6M2Q0dXh4NmZ5eTl6ZDVkeXIzeXhnenFxbmNkcW4=",
									"index": true
								}
							]
						},
						{
							"type": "wasm",
							"attributes": [
								{
									"key": "X2NvbnRyYWN0X2FkZHJlc3M=",
									"value": "c2VpMWdqcnJtZTIyY3loYTRodDJ4YXBuM2YwOHp6dzZ6M2Q0dXh4NmZ5eTl6ZDVkeXIzeXhnenFxbmNkcW4=",
									"index": true
								},
								{
									"key": "bWVzc2FnZS5tZXNzYWdl",
									"value": "MDEwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwM2U4MDY5Yjg4NTdmZWFiODE4NGZiNjg3ZjYzNDYxOGMwMzVkYWM0MzlkYzFhZWIzYjU1OThhMGYwMDAwMDAwMDAwMTAwMDFlZmUxOGUyYTMzNDIzNjZkNWQwODIzNzY2OTg5NTE0YzkwN2EyNDM2NjdkZDlmZjJhNGMzZmM0NmQyOGNhMjNmMDAwMTAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA=",
									"index": true
								},
								{
									"key": "bWVzc2FnZS5zZW5kZXI=",
									"value": "ODZjNWZkOTU3ZTJkYjgzODk1NTNlMTcyOGY5YzI3OTY0YjIyYTgxNTQwOTFjY2JhNTRkNzVmNGIxMGM2MWY1ZQ==",
									"index": true
								},
								{
									"key": "bWVzc2FnZS5jaGFpbl9pZA==",
									"value": "MzI=",
									"index": true
								},
								{
									"key": "bWVzc2FnZS5ub25jZQ==",
									"value": "MA==",
									"index": true
								},
								{
									"key": "bWVzc2FnZS5zZXF1ZW5jZQ==",
									"value": "MTU1NTc=",
									"index": true
								},
								{
									"key": "bWVzc2FnZS5ibG9ja190aW1l",
									"value": "MTY5NzgxMDU3Mg==",
									"index": true
								},
								{
									"key": "aXNfaWJj",
									"value": "dHJ1ZQ==",
									"index": true
								}
							]
						},
						{
							"type": "send_packet",
							"attributes": [
								{
									"key": "cGFja2V0X2RhdGE=",
									"value": "eyJwdWJsaXNoIjp7Im1zZyI6W3sia2V5IjoibWVzc2FnZS5tZXNzYWdlIiwidmFsdWUiOiIwMTAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAzZTgwNjliODg1N2ZlYWI4MTg0ZmI2ODdmNjM0NjE4YzAzNWRhYzQzOWRjMWFlYjNiNTU5OGEwZjAwMDAwMDAwMDAxMDAwMWVmZTE4ZTJhMzM0MjM2NmQ1ZDA4MjM3NjY5ODk1MTRjOTA3YTI0MzY2N2RkOWZmMmE0YzNmYzQ2ZDI4Y2EyM2YwMDAxMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMCJ9LHsia2V5IjoibWVzc2FnZS5zZW5kZXIiLCJ2YWx1ZSI6Ijg2YzVmZDk1N2UyZGI4Mzg5NTUzZTE3MjhmOWMyNzk2NGIyMmE4MTU0MDkxY2NiYTU0ZDc1ZjRiMTBjNjFmNWUifSx7ImtleSI6Im1lc3NhZ2UuY2hhaW5faWQiLCJ2YWx1ZSI6IjMyIn0seyJrZXkiOiJtZXNzYWdlLm5vbmNlIiwidmFsdWUiOiIwIn0seyJrZXkiOiJtZXNzYWdlLnNlcXVlbmNlIiwidmFsdWUiOiIxNTU1NyJ9LHsia2V5IjoibWVzc2FnZS5ibG9ja190aW1lIiwidmFsdWUiOiIxNjk3ODEwNTcyIn1dfX0=",
									"index": true
								},
								{
									"key": "cGFja2V0X2RhdGFfaGV4",
									"value": "N2IyMjcwNzU2MjZjNjk3MzY4MjIzYTdiMjI2ZDczNjcyMjNhNWI3YjIyNmI2NTc5MjIzYTIyNmQ2NTczNzM2MTY3NjUyZTZkNjU3MzczNjE2NzY1MjIyYzIyNzY2MTZjNzU2NTIyM2EyMjMwMzEzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMzNjUzODMwMzYzOTYyMzgzODM1Mzc2NjY1NjE2MjM4MzEzODM0NjY2MjM2MzgzNzY2MzYzMzM0MzYzMTM4NjMzMDMzMzU2NDYxNjMzNDMzMzk2NDYzMzE2MTY1NjIzMzYyMzUzNTM5Mzg2MTMwNjYzMDMwMzAzMDMwMzAzMDMwMzAzMDMxMzAzMDMwMzE2NTY2NjUzMTM4NjUzMjYxMzMzMzM0MzIzMzM2MzY2NDM1NjQzMDM4MzIzMzM3MzYzNjM5MzgzOTM1MzEzNDYzMzkzMDM3NjEzMjM0MzMzNjM2Mzc2NDY0Mzk2NjY2MzI2MTM0NjMzMzY2NjMzNDM2NjQzMjM4NjM2MTMyMzM2NjMwMzAzMDMxMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAzMDMwMzAyMjdkMmM3YjIyNmI2NTc5MjIzYTIyNmQ2NTczNzM2MTY3NjUyZTczNjU2ZTY0NjU3MjIyMmMyMjc2NjE2Yzc1NjUyMjNhMjIzODM2NjMzNTY2NjQzOTM1Mzc2NTMyNjQ2MjM4MzMzODM5MzUzNTMzNjUzMTM3MzIzODY2Mzk2MzMyMzczOTM2MzQ2MjMyMzI2MTM4MzEzNTM0MzAzOTMxNjM2MzYyNjEzNTM0NjQzNzM1NjYzNDYyMzEzMDYzMzYzMTY2MzU2NTIyN2QyYzdiMjI2YjY1NzkyMjNhMjI2ZDY1NzM3MzYxNjc2NTJlNjM2ODYxNjk2ZTVmNjk2NDIyMmMyMjc2NjE2Yzc1NjUyMjNhMjIzMzMyMjI3ZDJjN2IyMjZiNjU3OTIyM2EyMjZkNjU3MzczNjE2NzY1MmU2ZTZmNmU2MzY1MjIyYzIyNzY2MTZjNzU2NTIyM2EyMjMwMjI3ZDJjN2IyMjZiNjU3OTIyM2EyMjZkNjU3MzczNjE2NzY1MmU3MzY1NzE3NTY1NmU2MzY1MjIyYzIyNzY2MTZjNzU2NTIyM2EyMjMxMzUzNTM1MzcyMjdkMmM3YjIyNmI2NTc5MjIzYTIyNmQ2NTczNzM2MTY3NjUyZTYyNmM2ZjYzNmI1Zjc0Njk2ZDY1MjIyYzIyNzY2MTZjNzU2NTIyM2EyMjMxMzYzOTM3MzgzMTMwMzUzNzMyMjI3ZDVkN2Q3ZA==",
									"index": true
								},
								{
									"key": "cGFja2V0X3RpbWVvdXRfaGVpZ2h0",
									"value": "MC0w",
									"index": true
								},
								{
									"key": "cGFja2V0X3RpbWVvdXRfdGltZXN0YW1w",
									"value": "MTcyOTM0NjU3MjQwMTkzNTIzNg==",
									"index": true
								},
								{
									"key": "cGFja2V0X3NlcXVlbmNl",
									"value": "MTU1NjI=",
									"index": true
								},
								{
									"key": "cGFja2V0X3NyY19wb3J0",
									"value": "d2FzbS5zZWkxZ2pycm1lMjJjeWhhNGh0MnhhcG4zZjA4enp3NnozZDR1eHg2Znl5OXpkNWR5cjN5eGd6cXFuY2Rxbg==",
									"index": true
								},
								{
									"key": "cGFja2V0X3NyY19jaGFubmVs",
									"value": "Y2hhbm5lbC00",
									"index": true
								},
								{
									"key": "cGFja2V0X2RzdF9wb3J0",
									"value": "d2FzbS53b3JtaG9sZTF3a3d5MHhoODlrc2RnajlocjM0N2R5ZDJkdzd6ZXNtdHJ1ZTZrZnp5bWw0dmR0ejZlNXdzMnkwNTBy",
									"index": true
								},
								{
									"key": "cGFja2V0X2RzdF9jaGFubmVs",
									"value": "Y2hhbm5lbC0w",
									"index": true
								},
								{
									"key": "cGFja2V0X2NoYW5uZWxfb3JkZXJpbmc=",
									"value": "T1JERVJfVU5PUkRFUkVE",
									"index": true
								},
								{
									"key": "cGFja2V0X2Nvbm5lY3Rpb24=",
									"value": "Y29ubmVjdGlvbi02",
									"index": true
								}
							]
						}
					]
				},
				"tx": "CqgDCocDCiQvY29zbXdhc20ud2FzbS52MS5Nc2dFeGVjdXRlQ29udHJhY3QS3gIKKnNlaTE3ZHh1dmRmZ3h1MGdweW0zaHU4Z2xjY3Q5a2pjY240eHRkZmdmYxI+c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2wac3siY29udmVydF9hbmRfdHJhbnNmZXIiOnsicmVjaXBpZW50X2NoYWluIjoxLCJyZWNpcGllbnQiOiI3K0dPS2pOQ05tMWRDQ04yYVlsUlRKQjZKRFpuM1oveXBNUDhSdEtNb2o4PSIsImZlZSI6IjAifX0qewpzZmFjdG9yeS9zZWkxODlhZGd1YXd1Z2szZTU1em42M3o4cjlsbDI5eHJqd2NhNjM2cmE3djdneHV6bjk4c3h5cXd6dDQ3bC8zQXBMam92Z2tNVDRMV0FjcXlZRFBhTmlEU0ttdUpKZk1vbTE4RWQyOW8yNxIEMTAwMBIcV29ybWhvbGUgLSBJbml0aWF0ZSBUcmFuc2ZlchJnClAKRgofL2Nvc21vcy5jcnlwdG8uc2VjcDI1NmsxLlB1YktleRIjCiECqpEa+4OSbl7OBOxSFeMJnxZoPsR1kc32WZOujBb4Sk8SBAoCCAEYORITCg0KBHVzZWkSBTkwMzkzEPWVNxpAR0s5l65colOiSYxsQLEMtGW+iY3HCQ8GoDQmwkU5+MUSgB5WifNE8gc9gAnne8Pi7DNNJuZYNZWNAewST24kzg=="
			}
		],
		"total_count": "1"
	}
}
`

func TestXxx1(t *testing.T) {
	result, err := parseTxSearchResponse[seiTx]([]byte(jsonTxSearchResponse), &cosmosTxSearchParams{}, seiTxSearchExtractor)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "D97FD8EB0FAB7784A8A293A7FEF1F47FDE0C4375C254A19361E0F87CC01EF99A", result.TxHash)
	assert.Equal(t, "sei17dxuvdfgxu0gpym3hu8glcct9kjccn4xtdfgfc", result.Sender)
}
