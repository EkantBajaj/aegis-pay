from typing import Annotated, TypedDict, List, Union
from langchain_core.messages import BaseMessage
import operator

class AgentState(TypedDict):
    # The list of messages in the conversation (history)
    messages: Annotated[List[BaseMessage], operator.add]
    
    # Custom fields for our payment logic
    transaction_id: str
    amount: float
    currency: str
    original_error: str
    current_provider: str
    retry_count: int
    status: str  # "pending", "recovered", "failed"
    
    # The 'next' field tells LangGraph which node to visit next
    next: str
