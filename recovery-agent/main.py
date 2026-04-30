import os
import json
import time
import uuid
from confluent_kafka import Consumer, KafkaError
from dotenv import load_dotenv
from graph import create_recovery_graph
from langchain_core.messages import HumanMessage

# Load environment variables
load_dotenv(dotenv_path="../.env")

# Kafka Configuration
# We use a unique group.id each time to ensure we read from the beginning for this demo
unique_group_id = f"recovery-group-{uuid.uuid4().hex[:8]}"

KAFKA_CONF = {
    'bootstrap.servers': os.getenv("KAFKA_BROKERS", "localhost:19092"),
    'group.id': unique_group_id,
    'auto.offset.reset': 'earliest'
}

def process_transaction(transaction_data, graph):
    """
    Synchronously process the transaction through the LangGraph.
    """
    tx_id = transaction_data.get("transaction_id")
    request = transaction_data.get("request", {})
    
    print(f"\n[AGENT] >>> RECEIVED: {tx_id} | Amount: {request.get('amount')} <<<")
    
    initial_state = {
        "messages": [HumanMessage(content=f"Stripe failure detected. Starting recovery.")],
        "transaction_id": tx_id,
        "amount": float(request.get("amount", 0.0)),
        "currency": request.get("currency", "USD"),
        "original_error": transaction_data.get("error", "Timeout"),
        "current_provider": "Stripe",
        "retry_count": 0,
        "status": "pending",
        "next": "supervisor"
    }

    config = {"configurable": {"thread_id": tx_id}}

    try:
        print(f"  [AGENT] Calling Gemini Supervisor for decision...")
        # We use .stream() for synchronous execution
        for event in graph.stream(initial_state, config):
            for node, state in event.items():
                print(f"  [AGENT] Node [{node}] finished. State: {state.get('status', 'pending')}")
        
        print(f"[AGENT] >>> SUCCESS: Recovery process finished for {tx_id} <<<\n")
        
    except Exception as e:
        print(f"[AGENT ERROR] Graph execution failed: {e}")

def main():
    print("--- AEGIS-PAY AI AGENT: Initializing Graph & DB... ---")
    try:
        graph = create_recovery_graph()
    except Exception as e:
        print(f"CRITICAL ERROR during Graph Init: {e}")
        return
    
    print(f"--- AEGIS-PAY AI AGENT: Connecting to Kafka (Group: {unique_group_id})... ---")
    consumer = Consumer(KAFKA_CONF)
    consumer.subscribe(['failed-transactions'])

    print("--- AEGIS-PAY AI AGENT: Ready and Listening... ---")

    try:
        while True:
            msg = consumer.poll(1.0) # Poll every 1 second
            
            if msg is None:
                # No message? Just print a heartbeat every 10 empty polls to show we are alive
                continue
            
            if msg.error():
                print(f"Kafka Error: {msg.error()}")
                continue

            # We got a message!
            try:
                raw_val = msg.value().decode('utf-8')
                print(f"  [DEBUG] Raw Kafka Message Received: {raw_val[:100]}...")
                data = json.loads(raw_val)
                process_transaction(data, graph)
            except Exception as e:
                print(f"Error parsing/processing message: {e}")

    except KeyboardInterrupt:
        print("Stopping Agent...")
    finally:
        consumer.close()

if __name__ == "__main__":
    main()
