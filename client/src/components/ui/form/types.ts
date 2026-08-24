import {
  type ComponentPropsWithoutRef,
  type ElementType,
  type FormHTMLAttributes,
  type ReactNode,
} from "react";

export type FormProps = {
  className?: string;
  children: ReactNode;
  noValidate?: boolean;
} & FormHTMLAttributes<HTMLFormElement>;

export type FormGroupContextType = {
  controlId?: string;
  isInvalid: boolean;
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
  className?: string;
} & Omit<ComponentPropsWithoutRef<T>, "as" | "className">;

export type FormFeedbackProps = {
  type?: "invalid" | "valid";
  className?: string;
  children?: ReactNode;
};

export type FormTextProps = {
  className?: string;
  children?: ReactNode;
};
