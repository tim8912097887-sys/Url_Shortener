import { createContext, forwardRef, useContext, useId } from "react";

import { cn } from "../utils/cn";

import type {
  FormControlProps,
  FormFeedbackProps,
  FormGroupContextType,
  FormGroupProps,
  FormLabelProps,
  FormProps,
  FormTextProps,
} from "./types";

const FormGroupContext = createContext<FormGroupContextType>({
  controlId: undefined,
  isInvalid: false,
  feedbackId: undefined,
  textId: undefined,
});

export type FormComponent = typeof Form & {
  Group: typeof FormGroup;
  Label: typeof FormLabel;
  Control: typeof FormControl;
  Feedback: typeof FormFeedback;
  Text: typeof FormText;
};

function Form({ className, children, noValidate = true, ...props }: FormProps) {
  return (
    <form
      noValidate={noValidate}
      className={cn("space-y-5", className)}
      {...props}
    >
      {children}
    </form>
  );
}

function FormGroup({
  controlId,
  isInvalid = false,
  className,
  children,
}: FormGroupProps) {
  const generatedId = useId();

  const id = controlId ?? generatedId;
  const feedbackId = `${id}-feedback`;
  const textId = `${id}-description`;

  return (
    <FormGroupContext.Provider
      value={{
        controlId: id,
        isInvalid,
        feedbackId,
        textId,
      }}
    >
      <div className={cn("flex flex-col gap-1.5", className)}>{children}</div>
    </FormGroupContext.Provider>
  );
}

const FormLabel = forwardRef<HTMLLabelElement, FormLabelProps>(
  function FormLabel({ className, htmlFor, children, ...props }, ref) {
    const { controlId } = useContext(FormGroupContext);

    return (
      <label
        ref={ref}
        htmlFor={htmlFor ?? controlId}
        className={cn("text-sm font-medium text-slate-700", className)}
        {...props}
      >
        {children}
      </label>
    );
  },
);

const FormControl = forwardRef<
  HTMLInputElement | HTMLTextAreaElement,
  FormControlProps
>(function FormControl(
  {
    className,
    as: Component = "input",
    isInvalid,
    id,
    "aria-describedby": ariaDescribedBy,
    ...props
  },
  ref,
) {
  const ctx = useContext(FormGroupContext);

  const invalid = isInvalid ?? ctx.isInvalid;

  const describedBy =
    [ariaDescribedBy, ctx.textId, invalid ? ctx.feedbackId : undefined]
      .filter(Boolean)
      .join(" ") || undefined;

  return (
    <Component
      ref={ref as never}
      id={id ?? ctx.controlId}
      aria-invalid={invalid || undefined}
      aria-describedby={describedBy}
      className={cn(
        "w-full rounded-lg border bg-white px-3.5 py-2.5 text-sm text-slate-900 shadow-sm transition",
        "placeholder:text-slate-400",
        "focus:outline-none focus:ring-2 focus:ring-offset-0",
        invalid
          ? "border-red-400 focus:border-red-500 focus:ring-red-100"
          : "border-slate-300 focus:border-teal-500 focus:ring-teal-100",
        className,
      )}
      {...props}
    />
  );
});

function FormFeedback({
  type = "invalid",
  className,
  children,
}: FormFeedbackProps) {
  const { feedbackId } = useContext(FormGroupContext);

  if (children == null) return null;

  const isInvalid = type === "invalid";

  return (
    <p
      id={feedbackId}
      role={isInvalid ? "alert" : undefined}
      className={cn(
        "text-xs",
        isInvalid ? "text-red-600" : "text-teal-700",
        className,
      )}
    >
      {children}
    </p>
  );
}

function FormText({ className, children }: FormTextProps) {
  const { textId } = useContext(FormGroupContext);

  if (children == null) return null;

  return (
    <p id={textId} className={cn("text-xs text-slate-500", className)}>
      {children}
    </p>
  );
}

FormLabel.displayName = "Form.Label";
FormControl.displayName = "Form.Control";

const TypedForm = Form as FormComponent;

TypedForm.Group = FormGroup;
TypedForm.Label = FormLabel;
TypedForm.Control = FormControl;
TypedForm.Feedback = FormFeedback;
TypedForm.Text = FormText;

export default TypedForm;
