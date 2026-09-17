import type {
    FormFieldOption,
    FormFieldSummary,
    FormFieldValue,
} from "../../schema/applicants.types";

export function getOptionDisplayLabel(
    options: (string | FormFieldOption)[] | null | undefined,
    value: string
): string {
    if (!options || options.length === 0) return value;

    for (const opt of options) {
        if (typeof opt === "string") {
            if (opt === value) return opt;
        } else if (opt && typeof opt === "object") {
            if (String(opt.value) === value) {
                return opt.label || opt.value;
            }
        }
    }
    return value;
}

export function formatFieldValue(
    field: FormFieldSummary,
    val: FormFieldValue
): string {
    if (val === undefined || val === null || val === "") {
        return "-";
    }

    if (typeof val === "boolean") {
        return val ? "Yes" : "No";
    }

    const isOptionField =
        field.type === "radio" ||
        field.type === "checkbox" ||
        field.type === "select";

    if (Array.isArray(val)) {
        if (isOptionField && field.options) {
            return val
                .map((v) => getOptionDisplayLabel(field.options, String(v)))
                .join(", ");
        }
        return val.join(", ");
    }

    if (isOptionField && field.options) {
        return getOptionDisplayLabel(field.options, String(val));
    }

    if (field.type === "date") {
        const d = new Date(String(val));
        return isNaN(d.getTime()) ? String(val) : d.toLocaleDateString();
    }

    return String(val);
}
