import { describe, it, expect, vi } from "vitest";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { renderWithProviders } from "@/test-utils";
import { ContentTypeBuilder } from "../ContentTypeBuilder";
import type { FieldDefinition } from "@/types/cms";

const noop = () => Promise.resolve();

describe("ContentTypeBuilder — primitives", () => {
  it("renders a text input for type=text", () => {
    const schema: FieldDefinition[] = [{ name: "title", type: "text" }];
    renderWithProviders(
      <ContentTypeBuilder schema={schema} mutationFn={noop} />,
    );
    expect(screen.getByLabelText("title")).toBeInTheDocument();
  });

  it("renders a number input for type=number", () => {
    const schema: FieldDefinition[] = [{ name: "price", type: "number" }];
    renderWithProviders(
      <ContentTypeBuilder schema={schema} mutationFn={noop} />,
    );
    expect(screen.getByLabelText("price")).toBeInTheDocument();
  });

  it("renders a boolean switch for type=boolean", () => {
    const schema: FieldDefinition[] = [{ name: "active", type: "boolean" }];
    renderWithProviders(
      <ContentTypeBuilder schema={schema} mutationFn={noop} />,
    );
    expect(screen.getByRole("switch", { name: "active" })).toBeInTheDocument();
  });
});

describe("ContentTypeBuilder — layout", () => {
  it("renders a 2-col grid for type=layout", () => {
    const schema: FieldDefinition[] = [
      {
        name: "section",
        type: "layout",
        fields: [
          { name: "left", type: "text" },
          { name: "right", type: "text" },
        ],
      },
    ];
    const { container } = renderWithProviders(
      <ContentTypeBuilder schema={schema} mutationFn={noop} />,
    );
    const grid = container.querySelector(".grid");
    expect(grid).toBeInTheDocument();
    expect(screen.getByLabelText("left")).toBeInTheDocument();
    expect(screen.getByLabelText("right")).toBeInTheDocument();
  });
});

describe("ContentTypeBuilder — component", () => {
  it("renders a fieldset with legend for type=component", () => {
    const schema: FieldDefinition[] = [
      {
        name: "banner",
        type: "component",
        fields: [{ name: "title", type: "text" }],
      },
    ];
    renderWithProviders(
      <ContentTypeBuilder schema={schema} mutationFn={noop} />,
    );
    expect(screen.getByRole("group", { name: "banner" })).toBeInTheDocument();
  });

  it("uses dot-notation field names inside components", async () => {
    const onSubmit = vi.fn();
    const schema: FieldDefinition[] = [
      {
        name: "banner",
        type: "component",
        fields: [{ name: "title", type: "text" }],
      },
    ];
    renderWithProviders(
      <ContentTypeBuilder schema={schema} mutationFn={onSubmit} />,
    );
    const input = screen.getByLabelText("title");
    await userEvent.clear(input);
    await userEvent.type(input, "Hello");
    const btn = screen.getByRole("button", { name: /save/i });
    await userEvent.click(btn);
    // TanStack Query v5 calls mutationFn(variables, context); check first arg only
    const firstCallData = onSubmit.mock.calls[0][0] as Record<string, unknown>;
    expect(firstCallData).toMatchObject({ banner: { title: "Hello" } });
  });
});
