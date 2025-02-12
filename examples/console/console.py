import gradio as gr
import requests
import json

host = "127.1:8081"

prompt = {}

def config():
  reply = requests.get(f"http://{host}/config")
  return gr.Dropdown(choices=reply.json())

def prompt_llm(llm, role, task, reqs, inst, exinput, exreply, pdat, pctx):
  r = {}
  if len(role) != 0:
    r["stratum"] = role
  if len(task) != 0:
    r["task"] = task
  if len(reqs) != 0:
    r["requirements"] = txt2remark(reqs)
  if len(inst) != 0:
    r["instructions"] = txt2remark(inst)
  if len(exinput) != 0 and len(exreply) != 0:
    r["examples"] = txt2example(exinput, exreply)
  if len(pdat) != 0:
    r["input"] = txt2remark(pdat)
  if len(pctx) != 0:
    r["context"] = txt2remark(pctx)


  print(json.dumps(r, indent=2))

  reply = requests.post(f"http://localhost:8081/prompt?llm={llm}", json=r)
  return reply.json()

def txt2remark(txt):
  note = ""
  text = []
  for i, line in enumerate(txt.strip().split("\n")):
    if i == 0 and not line.startswith("* "):
      note = line
    if line.startswith("* "):
      text.append(line[2:])

  obj = {"text": text}
  if len(note) > 0:
    obj["note"] = note
  
  return obj

def txt2example(exin, exrp):
  exx = [] 
  ein = []
  for line in exin.strip().split("\n"):
    if line.startswith("* "):
      ein.append(line[2:])

  erp = []
  for line in exin.strip().split("\n"):
    if line.startswith("* "):
      ein.append(line[2:])

  for i in range(0, min(len(ein, erp))):
    obj = {"input": ein[i], "output": erp[i]}
    exx.append(obj)

  return exx


cdata = """
* DynamoDB has truly revolutionize our data management approach. DynamoDB scalability is a standout feature. The ability to seamlessly scale based on demand has been instrumental in accommodating our growing data requirements. The consistent, low-latency performance is commendable. Seamless scalability, coupled with lightning-fast performance. Large-scale datasets support without any issue. Flexibility and fully managed nature. As managed service, the DynamoDB simplifies the operational overhead of database management. Features such as automatic backups, security patching, and continuous monitoring contribute to a hassle-free experience.

* AWS DynamoDB is a super fast key-value pair DB store that is highly scalable and highly flexible. It provides read latency of single digit milliseconds and provides great integrations with other AWS services like AWS Lambda, AWS S3 via DynamoDB streams or Kinesis streams. It has helped make out microservices become highly scalable. It is easy to use and highly scalable and its SDK are present in all most all modern languages.

* DynamoDB has been a great DB solution for my team and I. The setup process is very easy to get started and iterate fast. We utilized it for direct lookups of pieces of data and it performs extremely well for that. It's a great NoSQL database option. The cost can be high for inserting large amounts of data and large-scale applications due to their read/write credit system. Limited query flexibility compared to traditional SQL databases. There is a learning curve for optimizing data models and queries.
""" 

def reset_to_L0():
  role = ""
  task = ""
  reqs = ""
  inst = ""
  exin = ""
  exrp = ""
  data = cdata
  ctxs = ""
  return gr.Textbox(role), gr.Textbox(task), gr.Textbox(reqs), gr.Textbox(inst), gr.Textbox(exin), gr.Textbox(exrp), gr.Textbox(data), gr.Textbox(ctxs)

def reset_to_L1():
  role = ""
  task = "Prepare a review by summarizing the reviewer comments."
  reqs = ""
  inst = ""
  exin = ""
  exrp = ""
  data = cdata
  ctxs = ""
  return gr.Textbox(role), gr.Textbox(task), gr.Textbox(reqs), gr.Textbox(inst), gr.Textbox(exin), gr.Textbox(exrp), gr.Textbox(data), gr.Textbox(ctxs)

def reset_to_L2():
  role = ""
  task = "Prepare a review by summarizing the following reviewer comments. The final output should highlight the core features of the technology, common strengths/weaknesses mentioned by multiple reviewers, suggestions for improvement."
  reqs = ""
  inst = ""
  exin = ""
  exrp = ""
  data = "The text for review is below:\n" + cdata
  ctxs = ""
  return gr.Textbox(role), gr.Textbox(task), gr.Textbox(reqs), gr.Textbox(inst), gr.Textbox(exin), gr.Textbox(exrp), gr.Textbox(data), gr.Textbox(ctxs)

def reset_to_L3():
  role = ""
  task = "Prepare a review by answering the following questions from the reviewer comments."
  reqs = """
* Based on the reviewer's comments, what are the core contributions made by the technology?

* What are the common strengths of this technology, as mentioned by multiple reviewers?

* What are the common weaknesses of this technology, as highlighted by multiple reviewers?

* What suggestions would you provide for improving this technology?

  """
  inst = ""
  exin = ""
  exrp = ""
  data = "The text for review is below:\n" + cdata
  ctxs = ""
  return gr.Textbox(role), gr.Textbox(task), gr.Textbox(reqs), gr.Textbox(inst), gr.Textbox(exin), gr.Textbox(exrp), gr.Textbox(data), gr.Textbox(ctxs)

def reset_to_L4():
  role = ""
  task = "Prepare a review by answering the following questions from the reviewer comments."
  reqs = """
* Based on the reviewer's comments, what are the core contributions made by the technology?

* What are the common strengths of this technology, as mentioned by multiple reviewers?

* What are the common weaknesses of this technology, as highlighted by multiple reviewers?

* What suggestions would you provide for improving this technology?

  """
  inst = """
* An output should highlight major strengths and issues mentioned by multiple reviewers, be less than 400 words in length, the response should be in English only.

  """
  exin = ""
  exrp = ""
  data = "The text for review is below:\n" + cdata
  ctxs = ""
  return gr.Textbox(role), gr.Textbox(task), gr.Textbox(reqs), gr.Textbox(inst), gr.Textbox(exin), gr.Textbox(exrp), gr.Textbox(data), gr.Textbox(ctxs)

def reset_to_L5():
  role = ""
  task = "Prepare a review by answering the following questions from the reviewer comments."
  reqs = """
* Based on the reviewer's comments, what are the core contributions made by the technology?

* What are the common strengths of this technology, as mentioned by multiple reviewers?

* What are the common weaknesses of this technology, as highlighted by multiple reviewers?

* What suggestions would you provide for improving this technology?

  """
  inst = """
* An output should highlight major strengths and issues mentioned by multiple reviewers, be less than 400 words in length, the response should be in English only.

* Use additional context to answer given questions.
  """
  exin = ""
  exrp = ""
  data = "The text for review is below:\n" + cdata
  ctxs = """
Below are additional context relevant to your goal task.

* the traditional data normalization techniques would not work with this database.

* the overall data design is based on understanding access patterns.

* the database is not designed for supporting SQL-like access.

* the first step in designing your DynamoDB application is to identify the specific query patterns that the system must satisfy.

  """
  return gr.Textbox(role), gr.Textbox(task), gr.Textbox(reqs), gr.Textbox(inst), gr.Textbox(exin), gr.Textbox(exrp), gr.Textbox(data), gr.Textbox(ctxs)

def reset_to_L6():
  role = ""
  task = "Prepare a review by answering the following questions from the reviewer comments."
  reqs = """
* Based on the reviewer's comments, what are the core contributions made by the technology?

* What are the common strengths of this technology, as mentioned by multiple reviewers?

* What are the common weaknesses of this technology, as highlighted by multiple reviewers?

* What suggestions would you provide for improving this technology?

  """
  inst = """
* An output should highlight major strengths and issues mentioned by multiple reviewers, be less than 400 words in length, the response should be in English only.

* Use additional context to answer given questions.

* Justify your response in detail by explaining why you made the choices you actually made.
  
  """
  exin = ""
  exrp = ""
  data = "The text for review is below:\n" + cdata
  ctxs = """
Below are additional context relevant to your goal task.

* the traditional data normalization techniques would not work with this database.

* the overall data design is based on understanding access patterns.

* the database is not designed for supporting SQL-like access.

* the first step in designing your DynamoDB application is to identify the specific query patterns that the system must satisfy.
  
  """
  return gr.Textbox(role), gr.Textbox(task), gr.Textbox(reqs), gr.Textbox(inst), gr.Textbox(exin), gr.Textbox(exrp), gr.Textbox(data), gr.Textbox(ctxs)


with gr.Blocks(theme=gr.themes.Monochrome()) as app:
  gr.Markdown(
    """
    # Chatter Console

    In the research community, there was an attempt for making [standardized taxonomy of prompts](https://aclanthology.org/2023.findings-emnlp.946.pdf) for large language models (LLMs) to solve complex tasks. It encourages the community to adopt the TELeR taxonomy to achieve meaningful comparisons among LLMs, facilitating more accurate conclusions and helping the community achieve consensus on state-of-the-art LLM performance more efficiently.

    This console if a front-end towards LLMs hosted by AWS BedRock, OpenAI and accessible through Golang [chatter](https://github.com/kshard/chatter) library:
    """)

  llm = gr.Dropdown(label="LLM", info="choose base LLM to be used here.", choices=[], interactive=True)

  gr.Markdown(
    """

    You can either craft the prompt following the template of choose one of presets:
    * L0: No directive, just data.
    * L1: Simple one sentence directive expressing the high-level goals.
    * L2: Multi-sentence (paragraph-style) directives expressing the high-level goals and the sub-tasks needs to be performed to achieve the goal.
    * L3: Complex (bulleted-style-list) directive expressing the high-level along with a detailed bulleted list of sub-tasks to be performed.
    * L4: Complex directives that includes following (i) description of the high-level goal (ii) a detailed bulleted list of sub-tasks (iii) a guideline on how LLM output will be evaluated, few examples.
    * L5: Complex directives that includes following (i) description of the high-level goal (ii) a detailed bulleted list of sub-tasks (iii) a guideline on how LLM output will be evaluated, few examples. (iv) additional relevant information gathered via retrieval-based techniques.
    * L6: Complex directives that includes following (i) description of the high-level goal (ii) a detailed bulleted list of sub-tasks (iii) a guideline on how LLM output will be evaluated, few examples. (iv) additional relevant information gathered via retrieval-based techniques. (v) an explicit statement asking LLM to explain its own output.

    """)
  with gr.Row():
    pL0 = gr.Button(value="L0", scale=1)
    pL1 = gr.Button(value="L1", scale=1)
    pL2 = gr.Button(value="L2", scale=1)
    pL3 = gr.Button(value="L3", scale=1)
    pL4 = gr.Button(value="L4", scale=1)
    pL5 = gr.Button(value="L5", scale=1)
    pL6 = gr.Button(value="L6", scale=1)


  with gr.Row(equal_height=True):
    role = gr.Textbox(lines=3, scale=2, interactive=True,
      label="Role",
      info="defines ground level constrain of the model behavior.",
      placeholder=(
        "Think about it as a cornerstone of the model behavior.\n"
        "Act as ...\n"
        "Your role is ...\n"
      ))
    task = gr.Textbox(lines=3, scale=3, interactive=True,
      label="Task (L1+)",
      info="the task is a summary of what you want the prompt to do.",
      placeholder=(
        "Multi-sentence (paragraph-style) directives expressing the high-level goals"
      )
    )

  with gr.Row(equal_height=True):
    reqs = gr.Textbox(lines=3, interactive=True,
      label="Requirements (L3+)",
      info="giving as much information to ensure LLM does not use any incorrect assumptions.",
      placeholder=(
        "* requirement 1\n"
        "* requirement 2\n"
        "* ...\n"
      ))

    inst = gr.Textbox(lines=3, interactive=True,
      label="Instructions (L4+)",
      info="instructions informs model how to complete the task.",
      placeholder=(
        "* instruction 1\n"
        "* instruction 2\n"
        "* ...\n"
      ))

  with gr.Row(equal_height=True):
    exinput = gr.Textbox(lines=3, interactive=True,
      label="Example Input (L4+)",
      info="example how to complete the task",
      placeholder=(
        "* example input 1\n"
        "* example input 2\n"
        "* ...\n"
      ))
    exreply = gr.Textbox(lines=3, interactive=True,
      label="Example Reply (L4+)",
      info="example how to complete the task",
      placeholder=(
        "* example output 1\n"
        "* example output 2\n"
        "* ...\n"
      ))

  with gr.Row(equal_height=True):
    pdat = gr.Textbox(lines=3, interactive=True,
      label="Input",
      info="input data required to complete the task.",
      placeholder=(
        "explain input on 1st row\n"
        "* input 1\n"
        "* input 2\n"
        "* ...\n"
      ))

  with gr.Row(equal_height=True):
    pctx = gr.Textbox(lines=3, interactive=True,
      label="Context (L5+)",
      info="additional information required to complete the task.",
      placeholder=(
        "explain context on 1st row\n"
        "* context input 1\n"
        "* context input 2\n"
        "* ...\n"
      ))

  with gr.Row(equal_height=True):
    reply = gr.Textbox(lines=3,
      label="Reply",
      info="response from LLM"
    )

  req = gr.Button(value="Prompt")
  req.click(fn=prompt_llm,
    inputs=[llm, role, task, reqs, inst, exinput, exreply, pdat, pctx],
    outputs=[reply]
  )

  pL0.click(fn=reset_to_L0, inputs=[], outputs=[role, task, reqs, inst, exinput, exreply, pdat, pctx])
  pL1.click(fn=reset_to_L1, inputs=[], outputs=[role, task, reqs, inst, exinput, exreply, pdat, pctx])
  pL2.click(fn=reset_to_L2, inputs=[], outputs=[role, task, reqs, inst, exinput, exreply, pdat, pctx])
  pL3.click(fn=reset_to_L3, inputs=[], outputs=[role, task, reqs, inst, exinput, exreply, pdat, pctx])
  pL4.click(fn=reset_to_L4, inputs=[], outputs=[role, task, reqs, inst, exinput, exreply, pdat, pctx])
  pL5.click(fn=reset_to_L5, inputs=[], outputs=[role, task, reqs, inst, exinput, exreply, pdat, pctx])
  pL6.click(fn=reset_to_L6, inputs=[], outputs=[role, task, reqs, inst, exinput, exreply, pdat, pctx])

  app.load(config, None, [llm])

if __name__ == "__main__":
    app.launch()
