import os
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_core.prompts import ChatPromptTemplate
from pydantic import BaseModel, Field
from typing import Literal
from state import AgentState

# 1. The Decision Structure
# We use Pydantic to force Gemini to return valid JSON that our code can read.
class Router(BaseModel):
    """Decide which worker to call next or finish the process."""
    next: Literal["AdyenProcessor", "PayPalProcessor", "FINISH"] = Field(
        description="The next node to execute. Choose FINISH if the task is complete or impossible."
    )
    reason: str = Field(description="The logical justification for this routing decision.")

def create_supervisor():
    # We use gemini-2.5-flash which is the current workhorse for the free tier in 2026
    llm = ChatGoogleGenerativeAI(
        model="gemini-2.5-flash", 
        google_api_key=os.getenv("GEMINI_API_KEY")
    )
    
    # The System Prompt defines the 'Business Logic' of our AI
    prompt = ChatPromptTemplate.from_messages([
        ("system", (
            "You are the Aegis-Pay Recovery Supervisor. Your mission: Recover failed payments.\n"
            "Context:\n"
            "- Original Error: {original_error}\n"
            "- Transaction: {amount} {currency}\n\n"
            "Routing Rules:\n"
            "1. If error is a timeout, try AdyenProcessor (highly reliable backup).\n"
            "2. If amount is > 1000, try AdyenProcessor first (better high-value handling).\n"
            "3. If Adyen fails or amount is small, try PayPalProcessor.\n"
            "4. If you have tried both and they failed, or the error is 'Fraud', choose FINISH."
        )),
        ("placeholder", "{messages}"),
    ])
    
    # This 'binds' the LLM to the Pydantic model. Gemini will now ONLY output JSON matching 'Router'.
    return prompt | llm.with_structured_output(Router)

def supervisor_node(state: AgentState):
    """
    The Entry Point Node. 
    It reads the current state and asks Gemini: 'Where do we go next?'
    """
    supervisor = create_supervisor()
    
    # We pass the relevant state fields to the prompt
    decision = supervisor.invoke({
        "original_error": state["original_error"],
        "amount": state["amount"],
        "currency": state["currency"],
        "messages": state["messages"]
    })
    
    # We log the reason for transparency (Senior move!)
    print(f"--- SUPERVISOR DECISION: {decision.next} ({decision.reason}) ---")
    
    # We return the update. In LangGraph, returning a dict merges it into the state.
    return {"next": decision.next}
