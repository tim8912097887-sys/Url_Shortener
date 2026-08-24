import { createContext, forwardRef, useContext } from "react";
import type {
  FormControlProps,
  FormFeedbackProps,
  FormGroupProps,
  FormLabelProps,
  FormProps,
  FormTextProps,
} from "./types";

type FormGroupContextType = {
  controlId?: string;
  isInvalid: boolean;
};

const FormGroupContext = createContext<FormGroupContextType>({
  controlId: undefined,
  isInvalid: false,
});

export type FormComponent = typeof Form & {
  Group: typeof FormGroup;
  Label: typeof FormLabel;
  Control: typeof FormControl;
  Feedback: typeof FormFeedback;
  Text: typeof FormText;
};

function Form({
  className = "",
  children,
  noValidate = true,
  ...props
}: FormProps) {
  return (
    <form
      noValidate={noValidate}
      className={`space-y-5 ${className}`}
      {...props}
    >
      {children}
    </form>
  );
}

function FormGroup({
  controlId,
  isInvalid = false,
  className = "",
  children,
}: FormGroupProps) {
  return (
    <FormGroupContext.Provider value={{ controlId, isInvalid }}>
      <div className={`flex flex-col gap-1.5 ${className}`}>{children}</div>
    </FormGroupContext.Provider>
  );
}

const FormLabel = forwardRef<HTMLLabelElement, FormLabelProps>(
  function FormLabel({ className = "", children, ...props }, ref) {
    const { controlId } = useContext(FormGroupContext);

    return (
      <label
        ref={ref}
        htmlFor={controlId}
        className={`text-sm font-medium text-slate-700 ${className}`}
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
>(function FormControl({ className = "", as, isInvalid, ...props }, ref) {
  const Component = as ?? "input";
  const ctx = useContext(FormGroupContext);
  const invalid = isInvalid ?? ctx.isInvalid;

  return (
    <Component
      ref={ref as never}
      id={ctx.controlId}
      aria-invalid={invalid || undefined}
      className={[
        "w-full rounded-lg border bg-white px-3.5 py-2.5 text-sm text-slate-900 shadow-sm transition",
        "placeholder:text-slate-400",
        "focus:outline-none focus:ring-2 focus:ring-offset-0",
        invalid
          ? "border-red-400 focus:border-red-500 focus:ring-red-100"
          : "border-slate-300 focus:border-teal-500 focus:ring-teal-100",
        className,
      ].join(" ")}
      {...props}
    />
  );
});

function FormFeedback({
  type = "invalid",
  className = "",
  children,
}: FormFeedbackProps) {
  if (!children) return null;

  const isInvalidType = type === "invalid";

  return (
    <p
      role={isInvalidType ? "alert" : undefined}
      className={`text-xs ${
        isInvalidType ? "text-red-600" : "text-teal-700"
      } ${className}`}
    >
      {children}
    </p>
  );
}

function FormText({ className = "", children }: FormTextProps) {
  return <p className={`text-xs text-slate-500 ${className}`}>{children}</p>;
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
