import requests
from langchain_core.messages import HumanMessage
from state import AgentState

# Configuration for our Mock APIs (running in Docker)
MOCK_API_URL = "http://localhost:8081"

def adyen_worker(state: AgentState):
    """Attempt recovery using Adyen."""
    print(f"--- WORKER: Attempting Adyen recovery for {state['transaction_id']} ---")
    
    payload = {
        "amount": state["amount"],
        "currency": state["currency"],
        "user_id": "ai_recovery_agent"
    }
    
    try:
        # We call our Python Mock service
        resp = requests.post(f"{MOCK_API_URL}/adyen/v1/payments", json=payload, timeout=5)
        
        if resp.status_code == 200:
            msg = f"Adyen recovery successful. Ref: {resp.json().get('pspReference')}"
            return {
                "messages": [HumanMessage(content=msg)],
                "status": "recovered",
                "next": "FINISH" # In a supervisor pattern, we usually go back to supervisor, but we can shortcut if successful
            }
        else:
            msg = f"Adyen recovery failed with status {resp.status_code}."
            return {
                "messages": [HumanMessage(content=msg)],
                "status": "pending",
                "next": "supervisor" # Hand back to PM to decide next move
            }
            
    except Exception as e:
        return {
            "messages": [HumanMessage(content=f"Adyen connection error: {str(e)}")],
            "next": "supervisor"
        }

def paypal_worker(state: AgentState):
    """Attempt recovery using PayPal."""
    print(f"--- WORKER: Attempting PayPal recovery for {state['transaction_id']} ---")
    
    payload = {
        "amount": state["amount"],
        "currency": state["currency"],
        "user_id": "ai_recovery_agent"
    }
    
    try:
        resp = requests.post(f"{MOCK_API_URL}/paypal/v1/checkout", json=payload, timeout=5)
        
        if resp.status_code == 200:
            return {
                "messages": [HumanMessage(content="PayPal recovery successful.")],
                "status": "recovered",
                "next": "FINISH"
            }
        else:
            return {
                "messages": [HumanMessage(content=f"PayPal failed: {resp.text}")],
                "status": "failed",
                "next": "supervisor"
            }
    except Exception as e:
        return {
            "messages": [HumanMessage(content=f"PayPal error: {str(e)}")],
            "next": "supervisor"
        }
