import { render, screen } from "@testing-library/react";
import { ApiError } from "../api/client";
import { ErrorAlert } from "./Feedback";

test("shows a user-facing error and trace id", () => {
  render(
    <ErrorAlert
      error={new ApiError(409, "out_of_stock", "raw message", "request-123")}
    />,
  );
  expect(screen.getByText("库存不足，本次秒杀未成功。")).toBeInTheDocument();
  expect(screen.getByText("request-123")).toBeInTheDocument();
});
