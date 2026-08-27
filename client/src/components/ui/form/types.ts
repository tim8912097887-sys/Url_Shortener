import type {
  ComponentPropsWithoutRef,
  ElementType,
  FormHTMLAttributes,
  ReactNode,
} from "react";

export type FormProps = {
  children?: ReactNode;
} & FormHTMLAttributes<HTMLFormElement>;

export type FormGroupContextType = {
  controlId?: string;
  isInvalid: boolean;
  feedbackId?: string;
  textId?: string;
};

export type FormGroupProps = {
  controlId?: string;
  isInvalid?: boolean;
  className?: string;
  children?: ReactNode;
};

export type FormLabelProps = ComponentPropsWithoutRef<"label">;

export type FormControlProps<T extends ElementType = "input"> = {
  as?: T;
  isInvalid?: boolean;
} & Omit<ComponentPropsWithoutRef<T>, "as">;

export type FormFeedbackProps = {
  type?: "invalid" | "valid";
  className?: string;
  children?: ReactNode;
};

export type FormTextProps = {
  className?: string;
  children?: ReactNode;
};
