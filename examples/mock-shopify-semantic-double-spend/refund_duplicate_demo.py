#!/usr/bin/env python3
"""
Demonstrates a semantic double-spend in a mock e-commerce refund flow.

The first refund request returns a simulated 504 Gateway Timeout, but the
backend still records the refund. The agent assumes the refund failed and
blindly retries, creating a duplicate refund for the same order.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Dict, List


class GatewayTimeoutError(RuntimeError):
    """Raised when the mock API times out after committing state."""


@dataclass
class MockShopifyAPI:
    refunds: List[Dict[str, object]] = field(default_factory=list)
    _request_count: int = 0

    def refund_order(self, user_id: str, order_id: str, amount: int) -> Dict[str, object]:
        """Simulate a refund endpoint that times out after the first successful write."""
        self._request_count += 1

        refund = {
            "refund_id": f"refund_{len(self.refunds) + 1}",
            "user_id": user_id,
            "order_id": order_id,
            "amount": amount,
            "status": "processed",
            "attempt": self._request_count,
        }
        self.refunds.append(refund)

        print(
            f"[Mock API] Processed refund in backend storage: "
            f"{refund['refund_id']} for user #{user_id} on order {order_id}"
        )

        if self._request_count == 1:
            raise GatewayTimeoutError(
                "504 Gateway Timeout: response lost after refund was committed"
            )

        return refund


def agent_refund_user(api: MockShopifyAPI, user_id: str, order_id: str, amount: int) -> None:
    print(f'[Agent] Instruction received: "Refund user #{user_id}"')

    for attempt in (1, 2):
        print(f"[Agent] Attempt {attempt}: issuing refund for order {order_id}...")
        try:
            refund = api.refund_order(user_id=user_id, order_id=order_id, amount=amount)
            print(f"[Agent] Refund reported as successful: {refund['refund_id']}")
            return
        except GatewayTimeoutError as exc:
            print(f"[Agent] Saw error: {exc}")
            print("[Agent] Assuming the refund failed. Retrying now.")


def print_summary(api: MockShopifyAPI) -> None:
    print("\nFinal refund ledger:")
    for refund in api.refunds:
        print(
            f"  - {refund['refund_id']}: user #{refund['user_id']}, "
            f"order {refund['order_id']}, amount ${refund['amount'] / 100:.2f}, "
            f"attempt {refund['attempt']}"
        )

    if len(api.refunds) > 1:
        print("\nResult: duplicate refund created.")
        print(
            "Why this matters: without execution guards, an LLM can misread a timeout "
            "as a failed mutation and accidentally issue the same refund twice."
        )


if __name__ == "__main__":
    mock_api = MockShopifyAPI()
    agent_refund_user(
        api=mock_api,
        user_id="123",
        order_id="order_456",
        amount=5000,
    )
    print_summary(mock_api)
