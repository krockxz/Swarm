"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Plus, Loader2 } from "lucide-react";
import { apiClient } from "@/lib/api-client";
import { useToast } from "@/components/ui/use-toast";

const createMissionSchema = z.object({
  name: z.string().min(1, "Name is required"),
  target_url: z.string().url("Invalid URL"),
  num_agents: z.number().min(1, "At least 1 agent").max(1000, "Maximum 1000 agents"),
  goal: z.string().min(1, "Goal is required"),
  max_duration_seconds: z
    .number()
    .min(10, "Minimum 10 seconds")
    .max(3600, "Maximum 3600 seconds (1 hour)"),
  rate_limit_per_second: z
    .number()
    .min(0.1, "Minimum 0.1 requests per second")
    .max(1000, "Maximum 1000 requests per second"),
  initial_system_prompt: z
    .string()
    .default(
      "You are an autonomous web testing agent. Navigate websites efficiently to achieve the given goal."
    ),
});

type CreateMissionFormValues = z.infer<typeof createMissionSchema>;

interface MissionCreateFormProps {
  onSuccess?: (missionId: string) => void;
}

export function MissionCreateForm({ onSuccess }: MissionCreateFormProps) {
  const [open, setOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { toast } = useToast();

  const form = useForm<CreateMissionFormValues>({
    resolver: zodResolver(createMissionSchema),
    defaultValues: {
      name: "",
      target_url: "https://",
      num_agents: 5,
      goal: "",
      max_duration_seconds: 300,
      rate_limit_per_second: 2.0,
      initial_system_prompt:
        "You are an autonomous web testing agent. Navigate websites efficiently to achieve the given goal.",
    },
  });

  const onSubmit = async (values: CreateMissionFormValues) => {
    setIsSubmitting(true);
    try {
      const response = await apiClient.createMission(values);
      toast({
        title: "Mission created",
        description: `Mission "${values.name}" has been started.`,
      });
      setOpen(false);
      form.reset();
      onSuccess?.(response.mission_id);
    } catch (error) {
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "Failed to create mission",
        variant: "destructive",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          New Mission
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create New Mission</DialogTitle>
          <DialogDescription>
            Configure a new AI-powered swarm testing mission. Agents will
            navigate the target website autonomously.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Mission Name</FormLabel>
                  <FormControl>
                    <Input placeholder="E-commerce Product Search" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="target_url"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Target URL</FormLabel>
                  <FormControl>
                    <Input placeholder="https://example.com" {...field} />
                  </FormControl>
                  <FormDescription>
                    The starting URL for agents to begin navigation
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="goal"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Mission Goal</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Search for 'laptop' and navigate to product details page"
                      className="resize-none"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    The objective that agents should try to accomplish
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="num_agents"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Number of Agents</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min={1}
                        max={1000}
                        {...field}
                        onChange={(e) =>
                          field.onChange(parseInt(e.target.value) || 1)
                        }
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="max_duration_seconds"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Max Duration (seconds)</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min={10}
                        max={3600}
                        {...field}
                        onChange={(e) =>
                          field.onChange(parseInt(e.target.value) || 300)
                        }
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="rate_limit_per_second"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Rate Limit (requests/second)</FormLabel>
                  <FormControl>
                    <Input
                      type="number"
                      min={0.1}
                      max={1000}
                      step={0.1}
                      {...field}
                      onChange={(e) =>
                        field.onChange(parseFloat(e.target.value) || 1.0)
                      }
                    />
                  </FormControl>
                  <FormDescription>
                    Controls how fast agents can make requests
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="initial_system_prompt"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>System Prompt (Optional)</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Custom instructions for the AI agents..."
                      className="resize-none h-24"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    Custom system prompt to override default AI behavior
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
                disabled={isSubmitting}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Start Mission
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
