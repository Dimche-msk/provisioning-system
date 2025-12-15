import { c as create_ssr_component, v as validate_component, h as compute_rest_props, i as spread, j as escape_attribute_value, k as escape_object, b as subscribe, f as escape, e as each } from "../../../../chunks/ssr.js";
import { $ as $format } from "../../../../chunks/index2.js";
import { I as Input } from "../../../../chunks/input.js";
import { T as Table, a as Table_header, b as Table_row, c as Table_head, d as Table_body, e as Table_cell } from "../../../../chunks/table-row.js";
import "clsx";
import { c as cn, B as Button, C as Card, a as Card_content } from "../../../../chunks/card-content.js";
import { b as badgeVariants } from "../../../../chunks/index3.js";
import "../../../../chunks/Toaster.svelte_svelte_type_style_lang.js";
import { I as Icon } from "../../../../chunks/Icon.js";
import { C as Check, T as Triangle_alert } from "../../../../chunks/triangle-alert.js";
import { X } from "../../../../chunks/x.js";
const Arrow_left = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [["path", { "d": "m12 19-7-7 7-7" }], ["path", { "d": "M19 12H5" }]];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "arrow-left" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const Badge = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $$restProps = compute_rest_props($$props, ["variant", "class"]);
  let { variant = "default" } = $$props;
  let { class: className = void 0 } = $$props;
  if ($$props.variant === void 0 && $$bindings.variant && variant !== void 0) $$bindings.variant(variant);
  if ($$props.class === void 0 && $$bindings.class && className !== void 0) $$bindings.class(className);
  return `<div${spread(
    [
      {
        class: escape_attribute_value(cn(badgeVariants({ variant }), className))
      },
      escape_object($$restProps)
    ],
    {}
  )}>${slots.default ? slots.default({}) : ``}</div>`;
});
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $t, $$unsubscribe_t;
  $$unsubscribe_t = subscribe($format, (value) => $t = value);
  let processing = false;
  let stats = {
    total: 0,
    success: 0,
    error: 0
  };
  let validatedRows = [];
  $$unsubscribe_t();
  return `<div class="h-full flex flex-col p-6 space-y-4 overflow-hidden"><div class="flex justify-between items-center shrink-0"><div class="flex items-center gap-4">${validate_component(Button, "Button").$$render(
    $$result,
    {
      variant: "ghost",
      size: "icon",
      href: "/phones"
    },
    {},
    {
      default: () => {
        return `${validate_component(Arrow_left, "ArrowLeft").$$render($$result, { class: "h-4 w-4" }, {}, {})}`;
      }
    }
  )} <h1 class="text-3xl font-bold tracking-tight">${escape($t("phones.import") || "Import Phones")}</h1></div></div> ${validate_component(Card, "Card.Root").$$render($$result, {}, {}, {
    default: () => {
      return `${validate_component(Card_content, "Card.Content").$$render($$result, { class: "p-6 space-y-4" }, {}, {
        default: () => {
          return `<div class="flex items-center gap-4">${validate_component(Input, "Input").$$render(
            $$result,
            {
              type: "file",
              accept: ".xlsx, .xls",
              disabled: processing
            },
            {},
            {}
          )} ${validatedRows.length > 0 ? `${validate_component(Button, "Button").$$render($$result, { disabled: processing }, {}, {
            default: () => {
              return `${``}
                        Import Valid Rows`;
            }
          })}` : ``}</div> ${validatedRows.length > 0 ? `<div class="flex gap-4 text-sm">${validate_component(Badge, "Badge").$$render($$result, { variant: "outline" }, {}, {
            default: () => {
              return `Total: ${escape(stats.total)}`;
            }
          })} ${validate_component(Badge, "Badge").$$render(
            $$result,
            {
              variant: "default",
              class: "bg-green-500"
            },
            {},
            {
              default: () => {
                return `Success: ${escape(stats.success)}`;
              }
            }
          )} ${validate_component(Badge, "Badge").$$render($$result, { variant: "destructive" }, {}, {
            default: () => {
              return `Error: ${escape(stats.error)}`;
            }
          })}</div>` : ``}`;
        }
      })}`;
    }
  })} ${validatedRows.length > 0 ? `<div class="flex-1 border rounded-md overflow-hidden flex flex-col bg-background"><div class="flex-1 overflow-y-auto">${validate_component(Table, "Table.Root").$$render($$result, {}, {}, {
    default: () => {
      return `${validate_component(Table_header, "Table.Header").$$render($$result, { class: "sticky top-0 bg-background z-10" }, {}, {
        default: () => {
          return `${validate_component(Table_row, "Table.Row").$$render($$result, {}, {}, {
            default: () => {
              return `${validate_component(Table_head, "Table.Head").$$render($$result, {}, {}, {
                default: () => {
                  return `Status`;
                }
              })} ${validate_component(Table_head, "Table.Head").$$render($$result, {}, {}, {
                default: () => {
                  return `MAC`;
                }
              })} ${validate_component(Table_head, "Table.Head").$$render($$result, {}, {}, {
                default: () => {
                  return `Number`;
                }
              })} ${validate_component(Table_head, "Table.Head").$$render($$result, {}, {}, {
                default: () => {
                  return `Vendor/Model`;
                }
              })} ${validate_component(Table_head, "Table.Head").$$render($$result, {}, {}, {
                default: () => {
                  return `User`;
                }
              })} ${validate_component(Table_head, "Table.Head").$$render($$result, {}, {}, {
                default: () => {
                  return `Message`;
                }
              })} ${validate_component(Table_head, "Table.Head").$$render($$result, { class: "text-right" }, {}, {
                default: () => {
                  return `Actions`;
                }
              })}`;
            }
          })}`;
        }
      })} ${validate_component(Table_body, "Table.Body").$$render($$result, {}, {}, {
        default: () => {
          return `${each(validatedRows, (row) => {
            return `${validate_component(Table_row, "Table.Row").$$render($$result, {}, {}, {
              default: () => {
                return `${validate_component(Table_cell, "Table.Cell").$$render($$result, {}, {}, {
                  default: () => {
                    return `${row.status === "success" ? `${validate_component(Check, "Check").$$render($$result, { class: "h-4 w-4 text-green-500" }, {}, {})}` : `${row.status === "error" ? `${validate_component(X, "X").$$render($$result, { class: "h-4 w-4 text-red-500" }, {}, {})}` : `${row.status === "conflict" ? `${validate_component(Triangle_alert, "AlertTriangle").$$render($$result, { class: "h-4 w-4 text-yellow-500" }, {}, {})}` : `${validate_component(Badge, "Badge").$$render($$result, { variant: "secondary" }, {}, {
                      default: () => {
                        return `Ready`;
                      }
                    })}`}`}`} `;
                  }
                })} ${validate_component(Table_cell, "Table.Cell").$$render($$result, {}, {}, {
                  default: () => {
                    return `${escape(row.mac)}`;
                  }
                })} ${validate_component(Table_cell, "Table.Cell").$$render($$result, {}, {}, {
                  default: () => {
                    return `${escape(row.number)}`;
                  }
                })} ${validate_component(Table_cell, "Table.Cell").$$render($$result, {}, {}, {
                  default: () => {
                    return `${escape(row.vendor)} / ${escape(row.model)}`;
                  }
                })} ${validate_component(Table_cell, "Table.Cell").$$render($$result, {}, {}, {
                  default: () => {
                    return `${escape(row.user)}`;
                  }
                })} ${validate_component(Table_cell, "Table.Cell").$$render($$result, { class: "text-sm text-muted-foreground" }, {}, {
                  default: () => {
                    return `${escape(row.message)} `;
                  }
                })} ${validate_component(Table_cell, "Table.Cell").$$render($$result, { class: "text-right" }, {}, {
                  default: () => {
                    return `${row.status === "conflict" ? `${validate_component(Button, "Button").$$render($$result, { size: "sm", variant: "outline" }, {}, {
                      default: () => {
                        return `Overwrite
                                        `;
                      }
                    })}` : `${row.status === "valid" ? `${validate_component(Button, "Button").$$render($$result, { size: "sm", variant: "ghost" }, {}, {
                      default: () => {
                        return `Import
                                        `;
                      }
                    })}` : ``}`} `;
                  }
                })} `;
              }
            })}`;
          })}`;
        }
      })}`;
    }
  })}</div></div>` : ``}</div>`;
});
export {
  Page as default
};
