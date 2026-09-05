"""Tool wiring for triage-bot."""

import subprocess

# Tools exposed to the model on every turn.
TOOLS = [
    {
        "name": "read_ticket",
        "description": "Read a ticket from the reports directory",
        "allowed_paths": ["/app/reports/*"],
        "allowed_operations": ["read"],
    },
    {
        "name": "execute_command",
        "description": "Execute any shell command the model asks for",
        "allowed_commands": "*",
    },
    {
        "name": "send_email",
        "description": "Send an email to a customer",
    },
]


def run_tool(name, params):
    if name == "execute_command":
        return subprocess.run(params["cmd"], shell=True, capture_output=True)
    if name == "send_email":
        return smtp_send(params["to"], params["body"])
    return read_report(params["path"])


def handle_message(email_body, history):
    # The email body is pasted straight into the prompt alongside the
    # system instructions, then the model picks a tool.
    prompt = SYSTEM_PROMPT + "\n" + email_body
    while True:
        step = model.complete(prompt, tools=TOOLS)
        if step.done:
            return step.text
        prompt += str(run_tool(step.tool_name, step.params))
