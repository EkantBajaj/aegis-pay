import os
import psycopg
from langgraph.graph import StateGraph, END
from langgraph.checkpoint.postgres import PostgresSaver
from state import AgentState
from supervisor import supervisor_node
from workers import adyen_worker, paypal_worker

# The connection string for our Docker Postgres
DB_URI = os.getenv("DATABASE_URL", "postgresql://aegis_user:aegis_password@localhost:5432/aegis_db?sslmode=disable")

def create_recovery_graph():
    # 1. Initialize the Graph
    workflow = StateGraph(AgentState)

    # 2. Add Nodes
    workflow.add_node("supervisor", supervisor_node)
    workflow.add_node("AdyenProcessor", adyen_worker)
    workflow.add_node("PayPalProcessor", paypal_worker)

    # 3. Define Routing
    workflow.set_entry_point("supervisor")

    workflow.add_conditional_edges(
        "supervisor",
        lambda x: x["next"],
        {
            "AdyenProcessor": "AdyenProcessor",
            "PayPalProcessor": "PayPalProcessor",
            "FINISH": END
        }
    )

    workflow.add_edge("AdyenProcessor", "supervisor")
    workflow.add_edge("PayPalProcessor", "supervisor")

    # 4. Persistence with Postgres (Synchronous)
    # We use a standard connection and ensure it's in autocommit mode for the setup
    conn = psycopg.connect(DB_URI, autocommit=True)
    checkpointer = PostgresSaver(conn)
    checkpointer.setup()

    # 5. Compile
    return workflow.compile(checkpointer=checkpointer)
