import {createServer} from 'node:http'
import {readFile} from 'node:fs/promises'
import {ChatOpenAI} from '@langchain/openai'
import {HumanMessage, SystemMessage, AIMessage, ToolMessage} from '@langchain/core/messages'
import {Client} from '@modelcontextprotocol/sdk/client/index.js'
import {StreamableHTTPClientTransport} from '@modelcontextprotocol/sdk/client/streamableHttp.js'

const instruction = await readFile(new URL('./prompts/system.tr.txt',import.meta.url),'utf8')
let turns=0, failures=0
createServer((req,res)=>{if(req.url!=='/metrics'){res.writeHead(404).end();return}
  res.setHeader('Content-Type','text/plain; version=0.0.4');
  res.end('imavara_service_ready 1\nimavara_agent_turns_total '+turns+'\nimavara_agent_errors_total '+failures+'\n')
}).listen(9091,'0.0.0.0')
createServer(async(req,res)=>{
  if(req.method!=='POST'||req.url!=='/turn'){res.writeHead(404).end();return}
  if(!req.headers.cookie){res.writeHead(400).end();return}
  turns++
  const mcp=new Client({name:'imavara-agent',version:'0.1.0'})
  const emit=(event:object)=>res.write(JSON.stringify(event)+'\n')
  try{
    let raw=''
    for await(const chunk of req){raw+=chunk;if(raw.length>2_000_000){res.writeHead(413).end();return}}
    const input=JSON.parse(raw)
    await mcp.connect(new StreamableHTTPClientTransport(new URL(process.env.MCP_URL||'http://mcp:8081/mcp'),{requestInit:{headers:{cookie:req.headers.cookie,'X-Turn-ID':input.turn_id}}}))
    const {tools}=await mcp.listTools()
    const model=new ChatOpenAI({apiKey:process.env.OPENROUTER_API_KEY,configuration:{baseURL:'https://openrouter.ai/api/v1'},model:'meta/muse-ember',temperature:0.3,maxRetries:0,modelKwargs:{reasoning:{effort:'low',exclude:true}},streaming:true})
      .bindTools(tools.map(t=>({type:'function',function:{name:t.name,description:t.description,parameters:t.inputSchema}})),{tool_choice:'auto'})
    const messages=[new SystemMessage(instruction)]
    for(const m of input.messages){
      if(m.role==='user')messages.push(new HumanMessage(m.content))
      else messages.push(new AIMessage(m.content+(m.cards?.length?'\nKart sırası: '+JSON.stringify(m.cards):'')))
    }
    if(input.reference_asset_id)messages.push(new HumanMessage('Seçili referans: '+input.reference_asset_id))
    res.writeHead(200,{'Content-Type':'application/x-ndjson'})
    let previousText=false
    for(let step=0;step<8;step++){
      let startedText=false
      let answer
      for await(const part of await model.stream(messages)){
        answer=answer?answer.concat(part):part
        if(typeof part.content==='string'&&part.content){
          emit({type:'delta',text:(!startedText && previousText?'\n\n':'')+part.content})
          startedText=true
          previousText=true
        }
      }
      if(!answer)throw Error('empty response')
      messages.push(answer)
      if(!answer.tool_calls?.length){emit({type:'done'});return}
      for(const tool of answer.tool_calls){
        emit({type:'tool',text:tool.name==='search_images'?'Katalogda aranıyor…':tool.name})
        const result=await mcp.callTool({name:tool.name,arguments:tool.args})
        const content=result.content.filter(v=>v.type==='text').map(v=>v.text).join('\n')
        if(!result.isError && tool.name==='search_images'){
          const data=JSON.parse(content);emit({type:'cards',cards:data.nearest?[data.nearest]:data.items})
        }
        if(!result.isError && ['generate_image','get_image'].includes(tool.name)){
          const card=JSON.parse(content)
          if(card.id && card.preview_url)emit({type:'cards',cards:[card]})
        }
        messages.push(new ToolMessage({content,tool_call_id:tool.id}))
      }
    }
    emit({type:'error',text:'İşlem sınırına ulaşıldı.'})
  }catch{failures++;if(!res.headersSent)res.writeHead(503);emit({type:'error',text:'Agent şu anda yanıt veremiyor.'})}
  finally{res.end();await mcp.close().catch(()=>{})}
}).listen(8082,'0.0.0.0')
