import {
    AlignLeft,
    Calendar,
    CalendarDays,
    CalendarRange,
    CheckSquare,
    CircleDot,
    Clock,
    Hash,
    Link,
    Mail,
    ToggleLeft,
    Type,
    UploadCloud,
    type LucideIcon,
} from "lucide-react";

import type { FormField, FormFieldType } from "@/schema/forms.types";

export interface FieldTypeMeta {
    type: FormFieldType;
    label: string;
    description: string;
    icon: LucideIcon;
    defaultLabel: string;
    defaultKeyPrefix: string;
    defaultPlaceholder: string;
    hasOptions?: boolean;
}

export const FIELD_TYPE_METAS: Record<FormFieldType, FieldTypeMeta> = {
    text: {
        type: "text",
        label: "Text",
        description: "Single-line plain text input",
        icon: Type,
        defaultLabel: "Short Answer",
        defaultKeyPrefix: "text",
        defaultPlaceholder: "Type your answer here...",
    },
    textarea: {
        type: "textarea",
        label: "Text Area",
        description: "Multi-line text input for long answers",
        icon: AlignLeft,
        defaultLabel: "Long Answer",
        defaultKeyPrefix: "textarea",
        defaultPlaceholder: "Type your detailed response here...",
    },
    email: {
        type: "email",
        label: "Email",
        description: "Email address with validation",
        icon: Mail,
        defaultLabel: "Email Address",
        defaultKeyPrefix: "email",
        defaultPlaceholder: "you@example.com",
    },
    number: {
        type: "number",
        label: "Number",
        description: "Numerical digits and counters",
        icon: Hash,
        defaultLabel: "Number",
        defaultKeyPrefix: "number",
        defaultPlaceholder: "0",
    },
    url: {
        type: "url",
        label: "URL",
        description: "Web page link with protocol validation",
        icon: Link,
        defaultLabel: "Website URL",
        defaultKeyPrefix: "url",
        defaultPlaceholder: "https://",
    },
    checkbox: {
        type: "checkbox",
        label: "Checkbox",
        description: "Multi-select option check list",
        icon: CheckSquare,
        defaultLabel: "Choose Options",
        defaultKeyPrefix: "checkbox",
        defaultPlaceholder: "",
        hasOptions: true,
    },
    radio: {
        type: "radio",
        label: "Radio Group",
        description: "Single choice option list",
        icon: CircleDot,
        defaultLabel: "Select One",
        defaultKeyPrefix: "radio",
        defaultPlaceholder: "",
        hasOptions: true,
    },
    switch: {
        type: "switch",
        label: "Switch",
        description: "Binary on/off or yes/no toggle",
        icon: ToggleLeft,
        defaultLabel: "Enable Option",
        defaultKeyPrefix: "switch",
        defaultPlaceholder: "",
    },
    date: {
        type: "date",
        label: "Date",
        description: "Calendar date picker",
        icon: Calendar,
        defaultLabel: "Select Date",
        defaultKeyPrefix: "date",
        defaultPlaceholder: "YYYY-MM-DD",
    },
    time: {
        type: "time",
        label: "Time",
        description: "Time of day selector",
        icon: Clock,
        defaultLabel: "Select Time",
        defaultKeyPrefix: "time",
        defaultPlaceholder: "HH:MM",
    },
    month: {
        type: "month",
        label: "Month",
        description: "Month and year selector",
        icon: CalendarRange,
        defaultLabel: "Select Month",
        defaultKeyPrefix: "month",
        defaultPlaceholder: "YYYY-MM",
    },
    week: {
        type: "week",
        label: "Week",
        description: "Week of the year selector",
        icon: CalendarDays,
        defaultLabel: "Select Week",
        defaultKeyPrefix: "week",
        defaultPlaceholder: "YYYY-Www",
    },
    file: {
        type: "file",
        label: "File Upload",
        description: "File attachment and upload",
        icon: UploadCloud,
        defaultLabel: "Upload Document",
        defaultKeyPrefix: "file",
        defaultPlaceholder: "Upload your document",
    },
};

export function createNewField(
    type: FormFieldType,
    existingCount: number = 0
): FormField {
    const meta = FIELD_TYPE_METAS[type];
    const fieldId = `field_${crypto.randomUUID().slice(0, 8)}`;
    const key = `${meta.defaultKeyPrefix}_${existingCount + 1}`;

    const defaultOptions = meta.hasOptions
        ? [
            { id: "opt_1", label: "Option 1", value: "option_1" },
            { id: "opt_2", label: "Option 2", value: "option_2" },
        ]
        : [];

    return {
        id: fieldId,
        key,
        label: meta.defaultLabel,
        type,
        required: false,
        placeholder: meta.defaultPlaceholder,
        is_system: false,
        options: defaultOptions,
        validation:
            type === "file"
                ? { accept: ".pdf, .png, .jpg, .jpeg", max_file_size_mb: 10 }
                : null,
    };
}
