import asyncio
from fastapi import FastAPI, Header, Response, HTTPException
from pydantic import BaseModel
from typing import Optional

app = FastAPI(title="Aegis-Pay Mock Providers")

class ChargeRequest(BaseModel):
    amount: float
    currency: str
    user_id: str

@app.post("/stripe/v1/charges")
async def stripe_charge(request: ChargeRequest):
    """
    Stripe Mock: Fails with 504 Timeout for high amounts
    """
    if request.amount > 1000:
        # Simulate a hanging connection
        await asyncio.sleep(10)
        return {"error": "request_timeout"}
    
    return {
        "id": "ch_stripe_success",
        "status": "succeeded",
        "amount": request.amount
    }

@app.post("/adyen/v1/payments")
async def adyen_charge(request: ChargeRequest, x_simulation: Optional[str] = Header(None)):
    """
    Adyen Mock: Fails with 403 based on a simulation header
    """
    if x_simulation == "blocked":
        return Response(content='{"error": "refused", "reason": "restricted_region"}', 
                        status_code=403, 
                        media_type="application/json")

    return {
        "pspReference": "ADYEN12345",
        "resultCode": "Authorised"
    }

@app.post("/paypal/v1/checkout")
async def paypal_charge(request: ChargeRequest):
    """
    PayPal Mock: Fails with 402 for specific user emails
    """
    if "low-balance" in request.user_id:
        raise HTTPException(status_code=402, detail="insufficient_funds")

    return {
        "id": "PAYPAL-OK-999",
        "state": "approved"
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8081)
